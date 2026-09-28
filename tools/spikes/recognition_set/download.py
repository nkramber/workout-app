#!/usr/bin/env python3
"""Fill the local cache with each image of the manifest, and compare each SHA-256 (D-97).

The cache is `.local/recognition_set/images/` at the repository root, and Git
ignores it. The script uses the network and makes no paid call. It keeps a
cached file when its hash agrees, so a second run downloads nothing.

    python3 tools/spikes/recognition_set/download.py            # fill the cache
    python3 tools/spikes/recognition_set/download.py --offline  # check the cache only

The exit code is 0 when each image of the manifest is in the cache with the
hash of the manifest, and 1 otherwise.
"""
import argparse
import hashlib
import os
import sys
import time
import urllib.error
import urllib.request

import manifest_check

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(manifest_check.HERE)))
CACHE = os.path.join(ROOT, ".local", "recognition_set")
# Wikimedia asks each client for a User-Agent with a contact.
USER_AGENT = "gym-route-recognition-test-set/1 (https://github.com/nkramber/workout-app)"
RETRIES = 4
# Commons answers 429 to a fast client. One file each 2 seconds passed in a
# test on 2026-09-28, and a 429 needs a long wait.
PAUSE_S = 2.0
RATE_LIMIT_WAIT_S = 30


def image_path(img, cache=CACHE):
    ext = os.path.splitext(img["file_url"])[1].lower()
    return os.path.join(cache, "images", img["image_id"] + ext)


def sha256_of(path):
    digest = hashlib.sha256()
    with open(path, "rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def cached_ok(img, cache=CACHE):
    path = image_path(img, cache)
    return os.path.exists(path) and sha256_of(path) == img["sha256"]


def fetch(url, opener=urllib.request.urlopen, sleep=time.sleep):
    """Return the bytes of the URL. Retry a rate limit or a server error."""
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    for attempt in range(RETRIES):
        try:
            with opener(request, timeout=60) as response:
                return response.read()
        except urllib.error.HTTPError as e:
            if e.code != 429 and e.code < 500 or attempt == RETRIES - 1:
                raise
            retry_after = e.headers.get("Retry-After") if e.headers else None
            if retry_after and retry_after.isdigit():
                sleep(float(retry_after))
            elif e.code == 429:
                sleep(RATE_LIMIT_WAIT_S * 2 ** attempt)
            else:
                sleep(2 ** (attempt + 1))
        except urllib.error.URLError:
            if attempt == RETRIES - 1:
                raise
            sleep(2 ** (attempt + 1))
    raise RuntimeError("unreachable")


def download(img, cache=CACHE, opener=urllib.request.urlopen, sleep=time.sleep):
    """Write the image to the cache. Return None, or the error text."""
    path = image_path(img, cache)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    try:
        data = fetch(img["file_url"], opener, sleep)
    except (urllib.error.URLError, OSError) as e:
        return f"{img['image_id']}: download failed: {e}"
    digest = hashlib.sha256(data).hexdigest()
    if digest != img["sha256"]:
        return f"{img['image_id']}: SHA-256 {digest} does not agree with the manifest"
    part = path + ".part"
    with open(part, "wb") as f:
        f.write(data)
    os.replace(part, path)
    return None


def run(images, cache=CACHE, offline=False, opener=urllib.request.urlopen, sleep=time.sleep, out=sys.stdout):
    kept = fetched = 0
    errors = []
    for img in images:
        if cached_ok(img, cache):
            kept += 1
            continue
        if offline:
            errors.append(f"{img['image_id']}: not in the cache, or the hash does not agree")
            continue
        error = download(img, cache, opener, sleep)
        if error:
            errors.append(error)
        else:
            fetched += 1
        sleep(PAUSE_S)
    for e in errors:
        print(e, file=out)
    print(f"{len(images)} images: {kept} in the cache, {fetched} downloaded, {len(errors)} errors", file=out)
    return errors


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--offline", action="store_true", help="check the cache only, with no network")
    args = parser.parse_args(argv)
    manifest = manifest_check.load("manifest.json")
    return 1 if run(manifest["images"], offline=args.offline) else 0


if __name__ == "__main__":
    sys.exit(main())
