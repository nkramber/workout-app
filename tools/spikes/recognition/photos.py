"""The photos of the recognition spike: the originals and the degraded
photos of the recognition test set, with the true machine type of each.

The harness sends each photo the way the app will send it (section 5.7 of
docs/research/platform-cloud-and-ai.md): the EXIF orientation applied, the
long side downscaled, JPEG, and no metadata. The preparation needs Pillow,
a local tool that CI does not install.
"""
import io
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SPIKES = os.path.dirname(HERE)
for folder in ("recognition_set", "luna_plan"):
    path = os.path.join(SPIKES, folder)
    if path not in sys.path:
        sys.path.append(path)

import download  # noqa: E402  (tools/spikes/recognition_set)
import manifest_check  # noqa: E402

GROUPS = ("original", "degraded")


def load_set():
    """Return (catalog, manifest) of the recognition test set."""
    return manifest_check.load("catalog.json"), manifest_check.load("manifest.json")


def photo_list(manifest, cache=download.CACHE):
    """Return one record for each photo: the originals first, then the
    degraded photos. A degraded photo takes the truth of its source."""
    by_id = {i["image_id"]: i for i in manifest["images"]}
    photos = []
    for img in manifest["images"]:
        photos.append({
            "photo_id": img["image_id"], "group": "original", "path": download.image_path(img, cache),
            "machine_type": img["machine_type"], "split": img["split"], "gym": img["gym"],
            "hard_negative": img["hard_negative"], "degrade": None,
        })
    for entry in manifest.get("degraded", []):
        source = by_id[entry["source"]]
        photos.append({
            "photo_id": entry["image_id"], "group": "degraded",
            "path": os.path.join(cache, "degraded", entry["image_id"] + ".jpg"),
            "machine_type": source["machine_type"], "split": source["split"], "gym": source["gym"],
            "hard_negative": source["hard_negative"], "degrade": f"{entry['kind']}-{entry['level']}",
            "source_id": source["image_id"],
        })
    return photos


def prepare(path, long_side, quality):
    """Return the JPEG bytes that the harness sends for one photo."""
    from PIL import Image, ImageOps

    with Image.open(path) as image:
        image = ImageOps.exif_transpose(image).convert("RGB")
    image.thumbnail((long_side, long_side))
    out = io.BytesIO()
    # Pillow writes no EXIF data unless the caller gives it.
    image.save(out, "JPEG", quality=quality)
    return out.getvalue()
