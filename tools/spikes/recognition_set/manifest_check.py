#!/usr/bin/env python3
"""Check the manifest of the recognition test set, with no network (D-95, D-97).

The check reads `manifest.json` and `catalog.json` of this folder. It refuses
each license outside D-95, and it checks the fields of each image, the split
by gym, and the design of section 6.6 of docs/research/platform-cloud-and-ai.md.
`make test` runs it through test_recognition_set.py. Run it alone with:

    python3 tools/spikes/recognition_set/manifest_check.py
"""
import json
import os
import re
import sys
from urllib.parse import unquote, urlparse

HERE = os.path.dirname(os.path.abspath(__file__))

# D-95: CC0, public domain, CC BY, CC BY-SA, and CC BY-NC. The codes are the
# license codes of the Commons extmetadata. A code with a jurisdiction port,
# such as cc-by-sa-3.0-de, is the same license. CC BY-NC-SA, CC BY-ND, GFDL,
# and "No restrictions" are outside D-95, so the check refuses them.
LICENSE = re.compile(r"^(cc0|pd|cc-by(-sa|-nc)?-[1-4]\.[05](-[a-z]{2})?)$")
LICENSE_PATH = {"cc-by": "/licenses/by/", "cc-by-sa": "/licenses/by-sa/", "cc-by-nc": "/licenses/by-nc/"}
SHA256 = re.compile(r"^[0-9a-f]{64}$")
IMAGE_ID = re.compile(r"^R\d{3}$")
DEGRADED_ID = re.compile(r"^D\d{3}$")
GYM_ID = re.compile(r"^G\d{2}$")
SPLITS = {"test", "calibration"}
GYM_BASES = {"named", "photographer_date"}
# D-103 and D-104: no face, or a face that D-104 permits for a catalog type
# with few face-free photos.
FACE_CHECKS = {"no_face", "face_d104"}
COMMONS_FILE = "https://commons.wikimedia.org/wiki/File:"
COMMONS_UPLOAD = "https://upload.wikimedia.org/wikipedia/commons/"
# D-105: Flickr through Openverse, for the types with too few Commons images.
# The Openverse record is the machine-readable license proof of the image.
FLICKR_PAGE = re.compile(r"^https://www\.flickr\.com/photos/[^/?#]+/\d+/?$")
FLICKR_FILE = re.compile(r"^https://live\.staticflickr\.com/[0-9a-z/_]+\.jpg$")
OPENVERSE_PROOF = re.compile(r"^https://api\.openverse\.org/v1/images/[0-9a-f-]{36}/$")
SOURCES = {"commons", "flickr"}
EXTENSIONS = (".jpg", ".jpeg", ".png")
DEGRADE_KINDS = {"blur", "low_light", "glare", "occlusion"}

IMAGE_FIELDS = {
    "image_id": str, "source": str, "title": str, "source_url": str, "file_url": str,
    "license_proof": (str, type(None)), "author": str,
    "license": str, "license_url": (str, type(None)), "sha256": str, "bytes": int,
    "machine_type": str, "subject": str, "gym": str, "split": str,
    "hard_negative": (str, type(None)), "face_check": str,
}

# The design limits. The roadmap recommends about 200 photos, so that the paid
# run of the recognition spike stays below its cap (D-94). D-103 accepts a
# smaller real count.
MIN_IMAGES, MAX_IMAGES = 100, 230
MAX_PHOTOS = 260
MIN_NONE, MIN_HARD_NEGATIVES = 20, 20
MIN_GYMS_EACH_TYPE = 2
CALIBRATION_SHARE = (0.2, 0.4)


def load(name, folder=HERE):
    with open(os.path.join(folder, name), encoding="utf-8") as f:
        return json.load(f)


def license_ok(code):
    """True when the license code is inside D-95."""
    return isinstance(code, str) and bool(LICENSE.match(code))


def license_url_ok(code, url):
    """True when the license URL names the same license as the code."""
    if code in ("cc0", "pd"):
        return url is None or url.startswith("https://") or url.startswith("http://")
    family = re.match(r"^(cc-by(?:-sa|-nc)?)-(\d\.\d)", code)
    if not family or not url:
        return False
    return LICENSE_PATH[family.group(1)] + family.group(2) in url


def check_image(img, catalog_ids, none_label, gyms):
    errors = []
    iid = img.get("image_id", "?")
    for field, kind in IMAGE_FIELDS.items():
        if field not in img:
            errors.append(f"{iid}: no field {field}")
        elif not isinstance(img[field], kind):
            errors.append(f"{iid}: {field} has the wrong type")
    extra = set(img) - set(IMAGE_FIELDS)
    if extra:
        errors.append(f"{iid}: unknown fields {sorted(extra)}")
    if errors:
        return errors
    if not IMAGE_ID.match(iid):
        errors.append(f"{iid}: the id is not R<3 digits>")
    if not license_ok(img["license"]):
        errors.append(f"{iid}: license {img['license']!r} is outside D-95")
    elif not license_url_ok(img["license"], img["license_url"]):
        errors.append(f"{iid}: license_url does not agree with {img['license']}")
    if not img["author"].strip():
        errors.append(f"{iid}: no author")
    if not SHA256.match(img["sha256"]):
        errors.append(f"{iid}: sha256 is not 64 lowercase hex digits")
    if img["bytes"] <= 0:
        errors.append(f"{iid}: bytes is not positive")
    errors += check_source(img)
    if img["machine_type"] != none_label and img["machine_type"] not in catalog_ids:
        errors.append(f"{iid}: machine_type {img['machine_type']!r} is not in the catalog")
    if img["hard_negative"] is not None and img["hard_negative"] not in catalog_ids:
        errors.append(f"{iid}: hard_negative {img['hard_negative']!r} is not in the catalog")
    if img["hard_negative"] is not None and img["hard_negative"] == img["machine_type"]:
        errors.append(f"{iid}: hard_negative names its own machine_type")
    if not img["subject"].strip():
        errors.append(f"{iid}: no subject")
    if img["split"] not in SPLITS:
        errors.append(f"{iid}: split {img['split']!r} is unknown")
    gym = gyms.get(img["gym"])
    if gym is None:
        errors.append(f"{iid}: gym {img['gym']!r} is not in the gym list")
    elif gym["split"] != img["split"]:
        errors.append(f"{iid}: split {img['split']} is not the split of {img['gym']}")
    if img["face_check"] not in FACE_CHECKS:
        errors.append(f"{iid}: face_check is not one of {sorted(FACE_CHECKS)}")
    elif img["face_check"] == "face_d104" and img["machine_type"] not in catalog_ids:
        errors.append(f"{iid}: face_d104 is only for a type of the catalog")
    return errors


def check_source(img):
    """The URL rules of each source (D-102, D-105)."""
    iid = img["image_id"]
    if img["source"] not in SOURCES:
        return [f"{iid}: source {img['source']!r} is unknown"]
    errors = []
    if not img["file_url"].lower().endswith(EXTENSIONS) or urlparse(img["file_url"]).query:
        errors.append(f"{iid}: file_url is not a plain JPEG or PNG URL")
    if img["source"] == "commons":
        if not img["title"].startswith("File:"):
            errors.append(f"{iid}: the title does not start with File:")
        elif unquote(img["source_url"]) != COMMONS_FILE + img["title"][5:].replace(" ", "_"):
            errors.append(f"{iid}: source_url does not agree with the title")
        if not img["file_url"].startswith(COMMONS_UPLOAD):
            errors.append(f"{iid}: file_url is not a Commons upload URL")
        if img["license_proof"] is not None:
            errors.append(f"{iid}: a Commons image has its proof on the file page, so license_proof is null")
    else:
        if not img["title"].strip():
            errors.append(f"{iid}: no title")
        if not FLICKR_PAGE.match(img["source_url"]):
            errors.append(f"{iid}: source_url is not a Flickr photo page")
        if not FLICKR_FILE.match(img["file_url"]):
            errors.append(f"{iid}: file_url is not a Flickr static file")
        if not img["license_proof"] or not OPENVERSE_PROOF.match(img["license_proof"]):
            errors.append(f"{iid}: a Flickr image needs its Openverse record as license_proof (D-105)")
    return errors


def check_gyms(gym_list):
    errors, gyms = [], {}
    for g in gym_list:
        gid = g.get("gym_id", "?")
        if set(g) != {"gym_id", "basis", "label", "split"}:
            errors.append(f"{gid}: a gym has the fields gym_id, basis, label, and split")
            continue
        if not GYM_ID.match(gid):
            errors.append(f"{gid}: the id is not G<2 digits>")
        if gid in gyms:
            errors.append(f"{gid}: duplicate gym id")
        if g["basis"] not in GYM_BASES:
            errors.append(f"{gid}: basis {g['basis']!r} is unknown")
        if g["split"] not in SPLITS:
            errors.append(f"{gid}: split {g['split']!r} is unknown")
        gyms[gid] = g
    return errors, gyms


def check_degraded(entries, images):
    errors, seen = [], set()
    for d in entries:
        did = d.get("image_id", "?")
        if set(d) != {"image_id", "source", "kind", "level"}:
            errors.append(f"{did}: a degraded photo has the fields image_id, source, kind, and level")
            continue
        if not DEGRADED_ID.match(did) or did in seen:
            errors.append(f"{did}: the id is not a unique D<3 digits>")
        seen.add(did)
        if d["source"] not in images:
            errors.append(f"{did}: source {d['source']!r} is not an image of the manifest")
        if d["kind"] not in DEGRADE_KINDS:
            errors.append(f"{did}: kind {d['kind']!r} is unknown")
        if d["level"] not in (1, 2):
            errors.append(f"{did}: level is not 1 or 2")
    kinds = {d.get("kind") for d in entries}
    if entries and not DEGRADE_KINDS <= kinds:
        errors.append(f"degraded: no photo of the kinds {sorted(DEGRADE_KINDS - kinds)}")
    return errors


def check_design(images, catalog_ids, none_label, gyms, degraded, gaps):
    errors = []
    for type_id in set(gaps) - set(catalog_ids):
        errors.append(f"design: known gap {type_id!r} is not in the catalog")
    if not MIN_IMAGES <= len(images) <= MAX_IMAGES:
        errors.append(f"design: {len(images)} images, outside {MIN_IMAGES} to {MAX_IMAGES}")
    if len(images) + len(degraded) > MAX_PHOTOS:
        errors.append(f"design: {len(images) + len(degraded)} photos, more than {MAX_PHOTOS}")
    for type_id in catalog_ids:
        type_gyms = {i["gym"] for i in images if i["machine_type"] == type_id}
        short = len(type_gyms) < MIN_GYMS_EACH_TYPE
        if short and type_id not in gaps:
            errors.append(f"design: {type_id} has images from {len(type_gyms)} gyms, fewer than {MIN_GYMS_EACH_TYPE}")
        if not short and type_id in gaps:
            errors.append(f"design: {type_id} is a known gap, but it has images from {len(type_gyms)} gyms")
    none = [i for i in images if i["machine_type"] == none_label]
    if len(none) < MIN_NONE:
        errors.append(f"design: {len(none)} images outside the catalog, fewer than {MIN_NONE}")
    hard = [i for i in images if i["hard_negative"] is not None]
    if len(hard) < MIN_HARD_NEGATIVES:
        errors.append(f"design: {len(hard)} hard negatives, fewer than {MIN_HARD_NEGATIVES}")
    calibration = sum(1 for i in images if i["split"] == "calibration")
    low, high = CALIBRATION_SHARE
    if images and not low <= calibration / len(images) <= high:
        errors.append(f"design: the calibration split holds {calibration} of {len(images)} images")
    unused = set(gyms) - {i["gym"] for i in images}
    if unused:
        errors.append(f"design: gyms with no image {sorted(unused)}")
    return errors


def check(manifest, catalog):
    """Return the list of errors of the manifest. An empty list is a pass."""
    catalog_ids = [t["type_id"] for t in catalog["types"]]
    none_label = catalog["none_label"]
    errors = []
    if manifest.get("catalog_version") != catalog.get("version"):
        errors.append("manifest: catalog_version does not agree with catalog.json")
    gym_errors, gyms = check_gyms(manifest.get("gyms", []))
    errors += gym_errors
    images = manifest.get("images", [])
    for img in images:
        errors += check_image(img, set(catalog_ids), none_label, gyms)
    if errors:
        return errors
    for field in ("image_id", "title", "file_url", "sha256"):
        values = [i[field] for i in images]
        duplicates = sorted({v for v in values if values.count(v) > 1})
        if duplicates:
            errors.append(f"manifest: duplicate {field} {duplicates[:3]}")
    by_id = {i["image_id"]: i for i in images}
    degraded = manifest.get("degraded", [])
    errors += check_degraded(degraded, by_id)
    gaps = manifest.get("known_gaps", {})
    if not all(isinstance(v, str) and v.strip() for v in gaps.values()):
        errors.append("manifest: each known gap needs a reason")
    errors += check_design(images, catalog_ids, none_label, gyms, degraded, gaps)
    return errors


def main():
    errors = check(load("manifest.json"), load("catalog.json"))
    for e in errors:
        print(e)
    manifest = load("manifest.json")
    print(f"{len(manifest['images'])} images, {len(manifest.get('degraded', []))} degraded photos, {len(errors)} errors")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
