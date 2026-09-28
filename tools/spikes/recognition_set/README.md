# Recognition test set

This folder holds the recognition test set of work area 1.2 of `docs/roadmaps/high-level-roadmap.md`. The recognition spike sends each photo of this set to Luna, one photo at a time. The spike measures the correct, wrong, and abstained answers, and each wrong answer with a high stated confidence. The spike code stays outside the product code (D-100).

Git holds no image (D-97). Git holds a manifest with the source, the author, the license, and the SHA-256 of each image. A script downloads the images to a local cache that Git ignores.

## Files

| File | Content |
|---|---|
| `catalog.json` | The catalog shortlist: the machine types that Luna can select. Fixed-path resistance machines and cardio machines only (D-45). Each type has a distinguishing detail. |
| `manifest.json` | The gyms, the images, and the degraded photos of the set. |
| `manifest_check.py` | The free check of the manifest. `make test` runs it. |
| `download.py` | Fills the local cache with each image, and compares each SHA-256. |
| `degrade.py` | Makes the degraded photos in the local cache. It needs Pillow. |
| `sources.py` | Reads the license, the author, and the file of each image from its source: Wikimedia Commons, or Flickr through Openverse. |
| `test_recognition_set.py` | The unit tests. They use no network. |

## The manifest

Each image has these fields:

| Field | Content |
|---|---|
| `image_id` | `R` and three digits. |
| `source` | `commons` (D-102), or `flickr` for a type with too few Commons images (D-105). |
| `title`, `source_url` | The title and the page of the image at its source. |
| `file_url` | The original file, with no query. |
| `license_proof` | For a Flickr image, the Openverse record that gives its license. For a Commons image, `null`, because the file page gives the license. |
| `author`, `license`, `license_url` | The attribution of the image. The license code is the code of the Commons metadata. |
| `sha256`, `bytes` | The hash and the size of the original file. |
| `machine_type` | A `type_id` of `catalog.json`, or `none_of_these` for a machine outside the catalog. |
| `subject` | A short description of the photo. |
| `gym`, `split` | The gym key and the split of the image. |
| `hard_negative` | The catalog type that the image looks like, or `null`. |
| `face_check` | `no_face`, or `face_d104`. The author examined each image. `face_d104` marks a photo that shows the face of a person. D-104 permits it only for a catalog type with few face-free photos. |

The gym list gives each gym key a basis. The basis `named` means that the source names the gym. The basis `photographer_date` means that the source names no gym. The key then comes from the photographer and the upload event (D-103).

The manifest also holds `known_gaps`. Each entry names a catalog type with images from fewer than two gyms, and the reason. The type stays in the catalog, so Luna can still select it in a wrong answer. The check fails when a known gap has two gyms or more.

A degraded photo has a source image, a kind, and a level. The kinds are `blur`, `low_light`, `glare`, and `occlusion`. Level 1 is mild, and level 2 is strong. The degraded photo has the machine type, the gym, and the split of its source image.

## The design

The design follows section 6.6 of `docs/research/platform-cloud-and-ai.md`:

- Each catalog type has images from two gyms or more.
- Hard negatives: an image that looks like a different catalog type, such as a plate-loaded and a selectorized version of the same movement.
- Machines outside the catalog: cable stations, racks, outdoor gym equipment, and other machines. Luna must abstain on them.
- Degraded photos with blur, low light, glare, and partial occlusion.
- The split is by gym, not by photo. The `calibration` split holds the gyms for the thresholds, and the `test` split holds the other gyms.

## The licenses

The check accepts only the licenses of D-95: CC0, public domain, CC BY, CC BY-SA, and CC BY-NC. It refuses each other license, for example GFDL, CC BY-NC-SA, CC BY-ND, and "No restrictions".

## Commands

Each command is free. The download commands use the network, and they make no paid call.

1. Check the manifest with no network:

```bash
python3 tools/spikes/recognition_set/manifest_check.py
```

2. Fill the local cache. The script waits 2 seconds between two files, because Commons refuses a fast client with HTTP 429:

```bash
python3 tools/spikes/recognition_set/download.py
```

3. Make the degraded photos:

```bash
python3 tools/spikes/recognition_set/degrade.py
```

4. Compare the license and the author of each image with its source:

```bash
python3 tools/spikes/recognition_set/sources.py check
```

Many originals store their rotation in the EXIF orientation tag. The degrade script applies the tag. A tool that sends an original must also apply it, or send the file with its EXIF data.

The cache is `.local/recognition_set/`. The images are in `images/`, and the degraded photos are in `degraded/`.

A file at its source can get a new version. The hash of the new version does not agree with the manifest, so `download.py` gives an error for that image. Replace the image in the manifest, or remove it.
