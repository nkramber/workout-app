#!/usr/bin/env python3
"""Make the degraded photos of the manifest in the local cache.

Section 6.6 of docs/research/platform-cloud-and-ai.md asks for photos with
blur, low light, glare, and partial occlusion. Git holds no image (D-97), so
this script makes each degraded photo from its cached source image. Run
download.py first. The script needs Pillow, a local tool that CI does not
install:

    python3 tools/spikes/recognition_set/degrade.py

The output is `.local/recognition_set/degraded/<image id>.jpg`. The recipes
use no random value, so the same source and the same Pillow version give the
same photo.
"""
import os
import sys

import download
import manifest_check

MAX_SIDE = 2048
JPEG_QUALITY = 85

# Each recipe gives the parameters of one kind at level 1 (mild) and level 2
# (strong). A size is a fraction of the long side, or of the width and height.
RECIPES = {
    "blur": {1: {"radius": 0.004}, 2: {"radius": 0.008}},
    "low_light": {1: {"brightness": 0.35, "contrast": 0.8}, 2: {"brightness": 0.2, "contrast": 0.7}},
    "glare": {1: {"center": (0.65, 0.35), "radius": 0.35, "strength": 0.75}, 2: {"center": (0.55, 0.4), "radius": 0.5, "strength": 0.9}},
    "occlusion": {1: {"box": (0.0, 0.55, 0.45, 1.0)}, 2: {"box": (0.0, 0.4, 0.6, 1.0)}},
}
OCCLUDER_GRAY = 96


def recipe(kind, level):
    return RECIPES[kind][level]


def output_path(entry, cache=download.CACHE):
    return os.path.join(cache, "degraded", entry["image_id"] + ".jpg")


def apply(image, kind, level):
    """Return a degraded copy of a Pillow image."""
    from PIL import Image, ImageDraw, ImageEnhance, ImageFilter, ImageOps

    # A phone photo often stores its rotation in the EXIF orientation tag.
    # Apply the tag, because the saved JPEG holds no EXIF data.
    image = ImageOps.exif_transpose(image).convert("RGB")
    image.thumbnail((MAX_SIDE, MAX_SIDE))
    w, h = image.size
    p = recipe(kind, level)
    if kind == "blur":
        return image.filter(ImageFilter.GaussianBlur(p["radius"] * max(w, h)))
    if kind == "low_light":
        image = ImageEnhance.Brightness(image).enhance(p["brightness"])
        return ImageEnhance.Contrast(image).enhance(p["contrast"])
    if kind == "glare":
        size = max(1, int(p["radius"] * max(w, h) * 2))
        # radial_gradient is black at the center and white at the edge.
        mask = Image.radial_gradient("L").resize((size, size))
        mask = mask.point(lambda v: int(max(0, 255 - v * 2) * p["strength"]))
        full = Image.new("L", (w, h), 0)
        full.paste(mask, (int(p["center"][0] * w - size / 2), int(p["center"][1] * h - size / 2)))
        return Image.composite(Image.new("RGB", (w, h), (255, 255, 255)), image, full)
    if kind == "occlusion":
        x0, y0, x1, y1 = p["box"]
        ImageDraw.Draw(image).rectangle((x0 * w, y0 * h, x1 * w, y1 * h), fill=(OCCLUDER_GRAY,) * 3)
        return image
    raise ValueError(kind)


def run(manifest, cache=download.CACHE, out=sys.stdout):
    from PIL import Image

    by_id = {i["image_id"]: i for i in manifest["images"]}
    errors = []
    for entry in manifest.get("degraded", []):
        source = by_id[entry["source"]]
        if not download.cached_ok(source, cache):
            errors.append(f"{entry['image_id']}: source {source['image_id']} is not in the cache, run download.py")
            continue
        with Image.open(download.image_path(source, cache)) as image:
            result = apply(image, entry["kind"], entry["level"])
        path = output_path(entry, cache)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        result.save(path, "JPEG", quality=JPEG_QUALITY)
    for e in errors:
        print(e, file=out)
    print(f"{len(manifest.get('degraded', []))} degraded photos, {len(errors)} errors", file=out)
    return errors


def main():
    try:
        import PIL  # noqa: F401
    except ImportError:
        print("degrade.py needs Pillow: python3 -m pip install Pillow")
        return 1
    return 1 if run(manifest_check.load("manifest.json")) else 0


if __name__ == "__main__":
    sys.exit(main())
