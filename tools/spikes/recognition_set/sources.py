#!/usr/bin/env python3
"""Read the license, the author, and the file of each image from its source.

The sources are Wikimedia Commons (D-102) and Flickr through Openverse
(D-105). Two commands. Each one uses the network and makes no paid call.

    python3 tools/spikes/recognition_set/sources.py build <selection.json>
    python3 tools/spikes/recognition_set/sources.py check

`build` reads a selection: the manifest without the fields that the source
gives. For a Commons image, it reads the author, the license, and the URLs
from Commons. For a Flickr image, it reads them from the Openverse record,
which is the license proof. It then downloads each file, and adds the size
and the SHA-256. It refuses each license outside D-95. `check` reads each
source again, and gives each difference of the license, the author, or the
file.
"""
import hashlib
import html
import json
import os
import re
import sys
import time
import urllib.parse

import download
import manifest_check

API = "https://commons.wikimedia.org/w/api.php"
OPENVERSE = "https://api.openverse.org/v1/images/"
BATCH = 50
TAG = re.compile(r"<[^>]+>")


def clean_text(value):
    """Return the plain text of an extmetadata value, which can hold HTML."""
    return " ".join(html.unescape(TAG.sub(" ", value or "")).split())


def plain_url(url):
    """Remove the query that the API adds to a file URL."""
    return url.split("?", 1)[0] if url else url


def api_get(params, opener=None, sleep=time.sleep):
    params = dict(params, format="json", formatversion="2")
    kwargs = {"opener": opener} if opener else {}
    return json.loads(download.fetch(API + "?" + urllib.parse.urlencode(params), sleep=sleep, **kwargs))


def metadata(titles, opener=None, sleep=time.sleep):
    """Return {title: fields} for each title, from the imageinfo of Commons."""
    out = {}
    for start in range(0, len(titles), BATCH):
        batch = titles[start:start + BATCH]
        data = api_get({
            "action": "query", "titles": "|".join(batch), "prop": "imageinfo",
            "iiprop": "url|size|sha1|user|extmetadata",
            "iiextmetadatafilter": "License|LicenseUrl|Artist",
        }, opener, sleep)
        # The API gives each title in its normal form. Map it back.
        normal = {n["to"]: n["from"] for n in data["query"].get("normalized", [])}
        for page in data["query"]["pages"]:
            title = normal.get(page["title"], page["title"])
            info = (page.get("imageinfo") or [{}])[0]
            ext = info.get("extmetadata", {})
            out[title] = {
                "missing": "missing" in page or not info,
                "file_url": plain_url(info.get("url")),
                "source_url": plain_url(info.get("descriptionurl")),
                "bytes": info.get("size"),
                "sha1": info.get("sha1"),
                "license": (ext.get("License") or {}).get("value"),
                "license_url": (ext.get("LicenseUrl") or {}).get("value") or None,
                # A file with no Artist field names its uploader as the author.
                "author": clean_text((ext.get("Artist") or {}).get("value"))
                or (f"{info['user']} (uploader)" if info.get("user") else ""),
            }
    return out


def openverse_license(record):
    """Return the license code of an Openverse record, in the Commons form."""
    code, version = record.get("license"), record.get("license_version")
    if code == "cc0":
        return "cc0"
    if code == "pdm":
        return "pd"
    return f"cc-{code}-{version}" if code and version else None


def openverse(uuid, opener=None, sleep=time.sleep):
    """Return the fields of one Openverse record."""
    kwargs = {"opener": opener} if opener else {}
    record = json.loads(download.fetch(OPENVERSE + uuid + "/", sleep=sleep, **kwargs))
    return {
        "missing": record.get("id") != uuid,
        "title": record.get("title") or "",
        "source_url": record.get("foreign_landing_url"),
        "file_url": plain_url(record.get("url")),
        "license_proof": OPENVERSE + uuid + "/",
        "license": openverse_license(record),
        "license_url": record.get("license_url"),
        "author": clean_text(record.get("creator")),
    }


def cached_file(sel, meta, cache):
    """Return the bytes of a cached Commons file whose SHA-1 agrees with
    Commons, or None. A second build then downloads nothing again."""
    if not cache or sel["source"] != "commons":
        return None
    path = download.image_path({"image_id": sel["image_id"], "file_url": meta["file_url"]}, cache)
    if not os.path.exists(path):
        return None
    with open(path, "rb") as f:
        data = f.read()
    return data if hashlib.sha1(data).hexdigest() == meta["sha1"] else None


def build(selection, opener=None, sleep=time.sleep, cache=None):
    """Return the manifest and the errors. Download each file to hash it, and
    write it to the cache when a cache is given."""
    titles = [i["title"] for i in selection["images"] if i["source"] == "commons"]
    meta = metadata(titles, opener, sleep)
    images, errors = [], []
    for sel in selection["images"]:
        if sel["source"] == "commons":
            m = meta.get(sel["title"])
            if m and not m["missing"]:
                m = dict(m, title=sel["title"], license_proof=None)
        else:
            m = openverse(sel["openverse_id"], opener, sleep)
        if not m or m["missing"]:
            errors.append(f"{sel['image_id']}: {sel.get('title') or sel.get('openverse_id')} is not at the source")
            continue
        if not manifest_check.license_ok(m["license"]):
            errors.append(f"{sel['image_id']}: license {m['license']!r} is outside D-95")
            continue
        data = cached_file(sel, m, cache)
        if data is None:
            kwargs = {"opener": opener} if opener else {}
            data = download.fetch(m["file_url"], sleep=sleep, **kwargs)
            sleep(download.PAUSE_S)
        if sel["source"] == "commons" and hashlib.sha1(data).hexdigest() != m["sha1"]:
            errors.append(f"{sel['image_id']}: the file does not agree with the SHA-1 of Commons")
            continue
        img = {
            "image_id": sel["image_id"], "source": sel["source"], "title": m["title"],
            "source_url": m["source_url"], "file_url": m["file_url"],
            "license_proof": m["license_proof"],
            "author": m["author"], "license": m["license"], "license_url": m["license_url"],
            "sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data),
        }
        for field in ("machine_type", "subject", "gym", "split", "hard_negative", "face_check"):
            img[field] = sel[field]
        images.append(img)
        if cache:
            path = download.image_path(img, cache)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "wb") as f:
                f.write(data)
    manifest = {k: v for k, v in selection.items() if k != "images"}
    manifest["images"] = images
    return manifest, errors


def check(manifest, opener=None, sleep=time.sleep):
    """Return each difference between the manifest and its sources."""
    commons = [i["title"] for i in manifest["images"] if i["source"] == "commons"]
    meta = metadata(commons, opener, sleep)
    errors = []
    for img in manifest["images"]:
        if img["source"] == "commons":
            m = meta.get(img["title"])
            fields = ("license", "author", "file_url", "bytes")
        else:
            m = openverse(img["license_proof"].rstrip("/").rsplit("/", 1)[1], opener, sleep)
            fields = ("license", "author", "file_url")
        if not m or m["missing"]:
            errors.append(f"{img['image_id']}: {img['title']} is not at the source")
            continue
        for field in fields:
            if m[field] != img[field]:
                errors.append(f"{img['image_id']}: {field} at the source is {m[field]!r}, the manifest has {img[field]!r}")
        if not manifest_check.license_ok(m["license"]):
            errors.append(f"{img['image_id']}: license {m['license']!r} at the source is outside D-95")
    return errors


def main(argv):
    if argv[:1] == ["build"] and len(argv) == 2:
        with open(argv[1], encoding="utf-8") as f:
            manifest, errors = build(json.load(f), cache=download.CACHE)
        with open(os.path.join(manifest_check.HERE, "manifest.json"), "w", encoding="utf-8") as f:
            json.dump(manifest, f, indent=1, ensure_ascii=False)
            f.write("\n")
        print(f"wrote {len(manifest['images'])} images")
    elif argv == ["check"]:
        errors = check(manifest_check.load("manifest.json"))
        print(f"{len(errors)} differences with the sources")
    else:
        print(__doc__)
        return 2
    for e in errors:
        print(e)
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
