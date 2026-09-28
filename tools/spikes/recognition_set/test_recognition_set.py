"""Unit tests of the recognition test set. They use no network: the Commons
calls and the downloads go to a fake opener."""
import copy
import hashlib
import io
import json
import os
import tempfile
import unittest
import urllib.error
from unittest import mock

import degrade
import download
import manifest_check
import sources

CATALOG = manifest_check.load("catalog.json")
MANIFEST = manifest_check.load("manifest.json")
TYPE_IDS = [t["type_id"] for t in CATALOG["types"]]


class FakeResponse(io.BytesIO):
    def __enter__(self):
        return self

    def __exit__(self, *exc):
        self.close()


def fake_opener(routes, calls=None):
    """Return an opener that answers each URL from routes. A route value is
    bytes, or an exception to raise."""
    def opener(request, timeout=None):
        url = request.full_url
        if calls is not None:
            calls.append(url)
        answer = routes[url]
        if isinstance(answer, list):
            answer = answer.pop(0)
        if isinstance(answer, Exception):
            raise answer
        return FakeResponse(answer)
    return opener


def http_error(code):
    return urllib.error.HTTPError("https://x", code, "error", {}, None)


def image(**changes):
    data = b"image bytes"
    img = {
        "image_id": "R001", "source": "commons", "title": "File:Leg press.jpg", "license_proof": None,
        "source_url": "https://commons.wikimedia.org/wiki/File:Leg_press.jpg",
        "file_url": "https://upload.wikimedia.org/wikipedia/commons/a/ab/Leg_press.jpg",
        "author": "A. Author", "license": "cc-by-sa-4.0",
        "license_url": "https://creativecommons.org/licenses/by-sa/4.0",
        "sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data),
        "machine_type": "leg_press_plate", "subject": "A plate-loaded leg press.",
        "gym": "G01", "split": "test", "hard_negative": None, "face_check": "no_face",
    }
    img.update(changes)
    return img


GYMS = {"G01": {"gym_id": "G01", "basis": "named", "label": "A gym", "split": "test"}}


class CatalogTest(unittest.TestCase):
    def test_types_are_unique_and_complete(self):
        self.assertEqual(len(TYPE_IDS), len(set(TYPE_IDS)))
        for t in CATALOG["types"]:
            self.assertEqual(set(t), {"type_id", "name", "kind", "load", "detail"}, t["type_id"])
            self.assertIn(t["kind"], {"strength", "cardio"})
            self.assertIn(t["load"], {"selectorized", "plate_loaded", "either", "none"})
            self.assertEqual(t["load"] == "none", t["kind"] == "cardio", t["type_id"])
        self.assertNotIn(CATALOG["none_label"], TYPE_IDS)


class LicenseTest(unittest.TestCase):
    def test_d95_licenses_pass(self):
        for code in ("cc0", "pd", "cc-by-2.0", "cc-by-4.0", "cc-by-sa-3.0", "cc-by-sa-3.0-de", "cc-by-nc-4.0"):
            self.assertTrue(manifest_check.license_ok(code), code)

    def test_other_licenses_fail(self):
        for code in (None, "", "gfdl", "GFDL 1.2", "cc-by-nc-sa-4.0", "cc-by-nd-4.0", "cc-by-nc-nd-3.0",
                     "No restrictions", "fal", "cc-by-sa-4.0 ", "CC BY-SA 4.0", "cc-by-9.0"):
            self.assertFalse(manifest_check.license_ok(code), code)

    def test_only_issued_versions_and_ports_pass(self):
        # P2-3 of the review of #4: the versions and ports of the CC legal tools list.
        for code in ("cc-by-1.0", "cc-by-sa-2.1-es", "cc-by-nc-2.1-jp", "cc-by-2.5-scotland",
                     "cc-by-sa-3.0-igo", "cc-by-sa-2.0-uk", "cc-by-nc-3.0-us"):
            self.assertTrue(manifest_check.license_ok(code), code)
        for code in ("cc-by-4.5", "cc-by-sa-3.5", "cc-by-nc-1.5", "cc-by-sa-2.1", "cc-by-4.0-de",
                     "cc-by-3.0-zz", "cc-by-sa-2.0-igo", "cc-by-sa-1.0-de", "cc-by-5.0"):
            self.assertFalse(manifest_check.license_ok(code), code)

    def test_the_url_of_an_unissued_license_fails(self):
        ok = manifest_check.license_url_ok
        self.assertTrue(ok("cc-by-sa-2.1-es", "https://creativecommons.org/licenses/by-sa/2.1/es/"))
        self.assertFalse(ok("cc-by-4.5", "https://creativecommons.org/licenses/by/4.5"))
        self.assertFalse(ok("cc-by-sa-3.5", "https://creativecommons.org/licenses/by-sa/3.5/"))
        self.assertFalse(ok("cc-by-3.0-zz", "https://creativecommons.org/licenses/by/3.0/zz/"))

    def test_the_check_refuses_an_image_outside_d95(self):
        errors = manifest_check.check_image(image(license="gfdl"), set(TYPE_IDS), "none_of_these", GYMS)
        self.assertEqual(errors, ["R001: license 'gfdl' is outside D-95"])

    def test_the_license_url_must_name_the_same_license(self):
        self.assertTrue(manifest_check.license_url_ok("cc-by-2.0", "https://creativecommons.org/licenses/by/2.0"))
        self.assertFalse(manifest_check.license_url_ok("cc-by-2.0", "https://creativecommons.org/licenses/by-sa/2.0"))
        self.assertFalse(manifest_check.license_url_ok("cc-by-sa-4.0", None))
        self.assertTrue(manifest_check.license_url_ok("pd", None))

    def test_the_license_url_must_be_on_creativecommons_org(self):
        ok = manifest_check.license_url_ok
        self.assertTrue(ok("cc-by-sa-2.0", "https://creativecommons.org/licenses/by-sa/2.0/"))
        self.assertTrue(ok("cc0", "http://creativecommons.org/publicdomain/zero/1.0/deed.en"))
        self.assertTrue(ok("cc0", None))
        self.assertTrue(ok("cc-by-sa-2.0-de", "https://creativecommons.org/licenses/by-sa/2.0/de/deed.en"))
        self.assertTrue(ok("cc-by-4.0", "https://creativecommons.org/licenses/by/4.0/legalcode"))
        self.assertTrue(ok("pd", "https://creativecommons.org/publicdomain/mark/1.0/"))
        for code, url in (("cc-by-2.0", "https://example.org/licenses/by/2.0"),
                          ("cc-by-2.0", "https://creativecommons.org.example.org/licenses/by/2.0"),
                          ("cc0", "https://example.org/publicdomain/zero/1.0/"),
                          ("pd", "https://example.org/"),
                          ("cc0", "https://creativecommons.org/licenses/by/2.0/"),
                          ("cc-by-sa-2.0", "https://creativecommons.org/licenses/by-sa/2.0evil/"),
                          ("cc0", "https://creativecommons.org/publicdomain/zero/1.0-not-a-license/"),
                          ("pd", "https://creativecommons.org/publicdomain/zero/1.0/"),
                          ("cc-by-sa-2.0", "https://creativecommons.org/licenses/by-sa/2.0/de/"),
                          ("cc-by-sa-2.0-de", "https://creativecommons.org/licenses/by-sa/2.0/"),
                          ("cc-by-4.0", "https://creativecommons.org/licenses/by/4.0/deed.en/extra")):
            self.assertFalse(ok(code, url), (code, url))


class ImageCheckTest(unittest.TestCase):
    def errors(self, **changes):
        return manifest_check.check_image(image(**changes), set(TYPE_IDS), "none_of_these", GYMS)

    def test_a_good_image_passes(self):
        self.assertEqual(self.errors(), [])

    def test_each_rule(self):
        cases = {
            "sha256 is not 64": {"sha256": "ABC"},
            "source_url does not agree": {"source_url": "https://commons.wikimedia.org/wiki/File:Other.jpg"},
            "not a plain JPEG or PNG URL": {"file_url": "https://upload.wikimedia.org/wikipedia/commons/a/ab/Leg_press.jpg?utm_source=x"},
            "is not a Commons upload URL": {"file_url": "https://example.org/wikipedia/commons/a/ab/Leg_press.jpg"},
            "license_proof is null": {"license_proof": "https://api.openverse.org/v1/images/x/"},
            "source 'web' is unknown": {"source": "web"},
            "is not in the catalog": {"machine_type": "cable_crossover"},
            "names its own machine_type": {"hard_negative": "leg_press_plate"},
            "is not the split of G01": {"split": "calibration"},
            "is not in the gym list": {"gym": "G09"},
            "face_check is not one of": {"face_check": "unchecked"},
            "no author": {"author": " "},
            "unknown fields": {"extra": 1},
        }
        for text, changes in cases.items():
            errors = self.errors(**changes)
            self.assertTrue(any(text in e for e in errors), (text, errors))

    def test_a_flickr_image_needs_its_proof(self):
        flickr = dict(source="flickr", title="MT3209 Seated Abdominal Crunch",
                      source_url="https://www.flickr.com/photos/127311295@N06/41989122941",
                      file_url="https://live.staticflickr.com/874/41989122941_0123abcd_b.jpg",
                      license="cc-by-sa-2.0", license_url="https://creativecommons.org/licenses/by-sa/2.0/",
                      license_proof="https://api.openverse.org/v1/images/0f0e8c3a-1b2c-4d5e-8f90-a1b2c3d4e5f6/")
        self.assertEqual(self.errors(**flickr), [])
        errors = self.errors(**dict(flickr, license_proof=None))
        self.assertTrue(any("needs its Openverse record" in e for e in errors), errors)
        errors = self.errors(**dict(flickr, source_url="https://www.flickr.com/search/?q=gym"))
        self.assertTrue(any("not a Flickr photo page" in e for e in errors), errors)

    def test_a_face_needs_a_catalog_type(self):
        self.assertEqual(self.errors(face_check="face_d104"), [])
        self.assertEqual(self.errors(face_check="face_d104", machine_type="treadmill"), [])
        errors = self.errors(face_check="face_d104", machine_type="none_of_these")
        self.assertTrue(any("face_d104 is only for a type of the catalog" in e for e in errors), errors)

    def test_a_percent_encoded_source_url_passes(self):
        errors = self.errors(title="File:Gold's Gym.jpg", source_url="https://commons.wikimedia.org/wiki/File:Gold%27s_Gym.jpg")
        self.assertEqual(errors, [])

    def test_a_missing_field_fails(self):
        img = image()
        del img["license"]
        errors = manifest_check.check_image(img, set(TYPE_IDS), "none_of_these", GYMS)
        self.assertEqual(errors, ["R001: no field license"])


class ManifestTest(unittest.TestCase):
    """The committed manifest obeys every rule. This is the free check of make verify."""

    def test_the_manifest_passes(self):
        self.assertEqual(manifest_check.check(MANIFEST, CATALOG), [])

    def test_each_gym_is_in_one_split(self):
        splits = {}
        for img in MANIFEST["images"]:
            splits.setdefault(img["gym"], set()).add(img["split"])
        self.assertTrue(all(len(s) == 1 for s in splits.values()))

    def test_a_duplicate_hash_fails(self):
        manifest = copy.deepcopy(MANIFEST)
        manifest["images"][1]["sha256"] = manifest["images"][0]["sha256"]
        self.assertTrue(any("duplicate sha256" in e for e in manifest_check.check(manifest, CATALOG)))

    def test_a_short_set_fails(self):
        manifest = copy.deepcopy(MANIFEST)
        manifest["images"] = manifest["images"][:10]
        manifest["degraded"] = []
        errors = manifest_check.check(manifest, CATALOG)
        self.assertTrue(any("10 images, outside" in e for e in errors), errors)

    def test_a_known_gap_must_be_real(self):
        manifest = copy.deepcopy(MANIFEST)
        manifest["known_gaps"]["treadmill"] = "Not a gap."
        errors = manifest_check.check(manifest, CATALOG)
        self.assertTrue(any("treadmill is a known gap, but" in e for e in errors), errors)

    def test_a_gap_needs_a_reason(self):
        manifest = copy.deepcopy(MANIFEST)
        gap = next(iter(manifest["known_gaps"]))
        manifest["known_gaps"][gap] = " "
        self.assertIn("manifest: each known gap needs a reason", manifest_check.check(manifest, CATALOG))

    def test_a_type_from_one_gym_fails(self):
        manifest = copy.deepcopy(MANIFEST)
        for img in manifest["images"]:
            if img["machine_type"] == "treadmill":
                img["gym"] = next(i["gym"] for i in manifest["images"] if i["machine_type"] == "treadmill")
                img["split"] = next(g["split"] for g in manifest["gyms"] if g["gym_id"] == img["gym"])
        errors = manifest_check.check(manifest, CATALOG)
        self.assertTrue(any("treadmill has images from 1 gyms" in e for e in errors), errors)

    def test_a_degraded_photo_needs_a_known_source(self):
        manifest = copy.deepcopy(MANIFEST)
        manifest["degraded"][0]["source"] = "R999"
        self.assertTrue(any("R999" in e for e in manifest_check.check(manifest, CATALOG)))


class DownloadTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.cache = self.tmp.name
        self.img = image()
        self.sleeps = []

    def tearDown(self):
        self.tmp.cleanup()

    def run_download(self, routes, offline=False, calls=None):
        out = io.StringIO()
        errors = download.run([self.img], self.cache, offline, fake_opener(routes, calls), self.sleeps.append, out)
        return errors, out.getvalue()

    def test_download_fills_the_cache(self):
        errors, out = self.run_download({self.img["file_url"]: b"image bytes"})
        self.assertEqual(errors, [])
        self.assertIn("1 downloaded", out)
        self.assertTrue(download.cached_ok(self.img, self.cache))
        self.assertTrue(download.image_path(self.img, self.cache).endswith("images/R001.jpg"))

    def test_a_second_run_downloads_nothing(self):
        self.run_download({self.img["file_url"]: b"image bytes"})
        calls = []
        errors, out = self.run_download({}, calls=calls)
        self.assertEqual((errors, calls), ([], []))
        self.assertIn("1 in the cache", out)

    def test_a_wrong_hash_fails_and_writes_nothing(self):
        errors, _ = self.run_download({self.img["file_url"]: b"other bytes"})
        self.assertEqual(len(errors), 1)
        self.assertIn("does not agree with the manifest", errors[0])
        self.assertFalse(os.path.exists(download.image_path(self.img, self.cache)))

    def test_a_changed_cache_file_downloads_again(self):
        path = download.image_path(self.img, self.cache)
        os.makedirs(os.path.dirname(path))
        with open(path, "wb") as f:
            f.write(b"changed")
        errors, out = self.run_download({self.img["file_url"]: b"image bytes"})
        self.assertEqual(errors, [])
        self.assertIn("1 downloaded", out)

    def test_offline_reads_no_network(self):
        calls = []
        errors, _ = self.run_download({}, offline=True, calls=calls)
        self.assertEqual(calls, [])
        self.assertIn("not in the cache", errors[0])

    def test_a_rate_limit_retries(self):
        routes = {self.img["file_url"]: [http_error(429), http_error(503), b"image bytes"]}
        errors, _ = self.run_download(routes)
        self.assertEqual(errors, [])
        self.assertEqual(self.sleeps[:2], [download.RATE_LIMIT_WAIT_S, 4])

    def test_a_missing_file_fails_at_once(self):
        calls = []
        errors, _ = self.run_download({self.img["file_url"]: http_error(404)}, calls=calls)
        self.assertEqual(len(calls), 1)
        self.assertIn("download failed", errors[0])

    def test_the_cache_is_ignored_by_git(self):
        self.assertTrue(download.CACHE.endswith(os.path.join(".local", "recognition_set")))
        with open(os.path.join(download.ROOT, ".gitignore"), encoding="utf-8") as f:
            self.assertIn(".local/", f.read().split())


class SourcesTest(unittest.TestCase):
    def test_clean_text_removes_html(self):
        self.assertEqual(sources.clean_text('<a href="x">Jane&nbsp;Doe</a>\n (talk)'), "Jane Doe (talk)")
        self.assertEqual(sources.plain_url("https://u/x.jpg?utm_source=a"), "https://u/x.jpg")

    def api_answer(self, license_code, data):
        return json.dumps({"query": {"pages": [{"title": "File:Leg press.jpg", "imageinfo": [{
            "url": "https://upload.wikimedia.org/wikipedia/commons/a/ab/Leg_press.jpg?utm_source=c",
            "descriptionurl": "https://commons.wikimedia.org/wiki/File:Leg_press.jpg",
            "size": len(data), "sha1": hashlib.sha1(data).hexdigest(),
            "extmetadata": {"License": {"value": license_code},
                            "LicenseUrl": {"value": "https://creativecommons.org/licenses/by-sa/4.0"},
                            "Artist": {"value": "<span>A. Author</span>"}}}]}]}}).encode()

    def run_build(self, license_code, data=b"image bytes", cache=None, calls=None):
        selection = {"version": 1, "images": [{k: v for k, v in image().items()
                                               if k in ("image_id", "source", "title", "machine_type", "subject", "gym",
                                                        "split", "hard_negative", "face_check")}]}
        answer = self.api_answer(license_code, data)
        def opener(request, timeout=None):
            if calls is not None:
                calls.append(request.full_url)
            return FakeResponse(answer if request.full_url.startswith(sources.API) else data)
        return sources.build(selection, opener=opener, sleep=lambda s: None, cache=cache)

    def test_build_gives_the_manifest_fields(self):
        manifest, errors = self.run_build("cc-by-sa-4.0")
        self.assertEqual(errors, [])
        self.assertEqual(manifest["images"], [image()])

    def test_build_names_the_uploader_when_no_artist_exists(self):
        answer = json.loads(self.api_answer("pd", b"image bytes"))
        info = answer["query"]["pages"][0]["imageinfo"][0]
        del info["extmetadata"]["Artist"]
        info["user"] = "Uploader"
        meta = sources.metadata(["File:Leg press.jpg"], opener=lambda r, timeout=None: FakeResponse(json.dumps(answer).encode()),
                                sleep=lambda s: None)
        self.assertEqual(meta["File:Leg press.jpg"]["author"], "Uploader (uploader)")

    def test_a_second_build_reads_the_cache(self):
        with tempfile.TemporaryDirectory() as cache:
            self.run_build("cc-by-sa-4.0", cache=cache)
            calls = []
            manifest, errors = self.run_build("cc-by-sa-4.0", cache=cache, calls=calls)
            self.assertEqual(errors, [])
            self.assertTrue(all(u.startswith(sources.API) for u in calls), calls)

    def test_build_refuses_a_license_outside_d95(self):
        manifest, errors = self.run_build("gfdl")
        self.assertEqual(manifest["images"], [])
        self.assertEqual(errors, ["R001: license 'gfdl' is outside D-95"])

    def test_build_reads_a_flickr_image_from_openverse(self):
        uuid = "0f0e8c3a-1b2c-4d5e-8f90-a1b2c3d4e5f6"
        record = json.dumps({"id": uuid, "title": "MT3209 Seated Abdominal Crunch", "creator": "haswell.fitness",
                             "license": "by-sa", "license_version": "2.0",
                             "license_url": "https://creativecommons.org/licenses/by-sa/2.0/",
                             "url": "https://live.staticflickr.com/874/41989122941_0123abcd_b.jpg",
                             "foreign_landing_url": "https://www.flickr.com/photos/127311295@N06/41989122941"}).encode()
        selection = {"images": [{"image_id": "R002", "source": "flickr", "openverse_id": uuid, "machine_type": "ab_crunch",
                                 "subject": "Abdominal crunch machine.", "gym": "G01", "split": "test",
                                 "hard_negative": None, "face_check": "no_face"}]}
        opener = lambda request, timeout=None: FakeResponse(
            record if request.full_url.startswith(sources.OPENVERSE) else b"flickr bytes")
        manifest, errors = sources.build(selection, opener=opener, sleep=lambda s: None)
        self.assertEqual(errors, [])
        img = manifest["images"][0]
        self.assertEqual((img["license"], img["license_proof"]), ("cc-by-sa-2.0", sources.OPENVERSE + uuid + "/"))
        self.assertEqual(manifest_check.check_source(img), [])

    def test_openverse_license_codes(self):
        self.assertEqual(sources.openverse_license({"license": "by-nc", "license_version": "2.0"}), "cc-by-nc-2.0")
        self.assertEqual(sources.openverse_license({"license": "pdm", "license_version": "1.0"}), "pd")
        self.assertFalse(manifest_check.license_ok(sources.openverse_license({"license": "by-nc-sa", "license_version": "2.0"})))

    def test_check_gives_each_changed_url(self):
        answer = self.api_answer("cc-by-sa-4.0", b"image bytes")
        opener = lambda request, timeout=None: FakeResponse(answer)
        changes = {"source_url": "https://commons.wikimedia.org/wiki/File:Other.jpg",
                   "license_url": "https://creativecommons.org/licenses/by-sa/4.0/deed.fr"}
        for field, value in changes.items():
            errors = sources.check({"images": [image(**{field: value})]}, opener=opener, sleep=lambda s: None)
            self.assertTrue(any(f"{field} at the source" in e for e in errors), (field, errors))

    def test_check_gives_each_changed_flickr_url(self):
        uuid = "0f0e8c3a-1b2c-4d5e-8f90-a1b2c3d4e5f6"
        record = {"id": uuid, "title": "T", "creator": "A. Author", "license": "by-sa", "license_version": "2.0",
                  "license_url": "https://creativecommons.org/licenses/by-sa/2.0/",
                  "url": "https://live.staticflickr.com/874/41989122941_0123abcd_b.jpg",
                  "foreign_landing_url": "https://www.flickr.com/photos/127311295@N06/41989122941"}
        img = image(source="flickr", title="T", source_url=record["foreign_landing_url"], file_url=record["url"],
                    license="cc-by-sa-2.0", license_url=record["license_url"],
                    license_proof=sources.OPENVERSE + uuid + "/")
        opener = lambda request, timeout=None: FakeResponse(json.dumps(record).encode())
        self.assertEqual(sources.check({"images": [img]}, opener=opener, sleep=lambda s: None), [])
        changes = {"source_url": "https://www.flickr.com/photos/127311295@N06/1",
                   "license_url": "https://creativecommons.org/licenses/by-sa/2.0/deed.de"}
        for field, value in changes.items():
            errors = sources.check({"images": [dict(img, **{field: value})]}, opener=opener, sleep=lambda s: None)
            self.assertTrue(any(f"{field} at the source" in e for e in errors), (field, errors))

    def test_check_gives_a_changed_license(self):
        answer = self.api_answer("cc-by-nc-sa-4.0", b"image bytes")
        opener = lambda request, timeout=None: FakeResponse(answer)
        errors = sources.check({"images": [image()]}, opener=opener, sleep=lambda s: None)
        self.assertTrue(any("license at the source is 'cc-by-nc-sa-4.0'" in e for e in errors), errors)
        self.assertTrue(any("outside D-95" in e for e in errors), errors)


class DegradeTest(unittest.TestCase):
    def test_each_kind_has_two_levels(self):
        self.assertEqual(set(degrade.RECIPES), manifest_check.DEGRADE_KINDS)
        for kind, levels in degrade.RECIPES.items():
            self.assertEqual(set(levels), {1, 2}, kind)

    def test_level_two_is_stronger(self):
        r = degrade.recipe
        self.assertGreater(r("blur", 2)["radius"], r("blur", 1)["radius"])
        self.assertLess(r("low_light", 2)["brightness"], r("low_light", 1)["brightness"])
        self.assertGreater(r("glare", 2)["radius"], r("glare", 1)["radius"])
        area = lambda b: (b[2] - b[0]) * (b[3] - b[1])
        self.assertGreater(area(r("occlusion", 2)["box"]), area(r("occlusion", 1)["box"]))
        self.assertLess(area(r("occlusion", 2)["box"]), 0.5)

    def test_the_output_is_in_the_cache(self):
        path = degrade.output_path({"image_id": "D001"})
        self.assertTrue(path.startswith(download.CACHE))

    def test_apply_makes_a_changed_photo(self):
        try:
            from PIL import Image, ImageChops
        except ImportError:
            self.skipTest("Pillow is a local tool, and CI does not install it")
        source = Image.new("RGB", (400, 300), (120, 140, 160))
        for kind in degrade.RECIPES:
            for level in (1, 2):
                result = degrade.apply(source, kind, level)
                self.assertEqual(result.size, (400, 300))
                if kind != "blur":
                    self.assertIsNotNone(ImageChops.difference(source, result).getbbox(), (kind, level))
                self.assertEqual(degrade.apply(source, kind, level).tobytes(), result.tobytes())


class DegradeOrientationTest(unittest.TestCase):
    def test_apply_obeys_the_exif_orientation(self):
        try:
            from PIL import Image
        except ImportError:
            self.skipTest("Pillow is a local tool, and CI does not install it")
        source = Image.new("RGB", (400, 300), (120, 140, 160))
        exif = source.getexif()
        exif[0x0112] = 6  # rotate 90 degrees clockwise to show
        data = io.BytesIO()
        source.save(data, "JPEG", exif=exif.tobytes())
        data.seek(0)
        self.assertEqual(degrade.apply(Image.open(data), "blur", 1).size, (300, 400))


if __name__ == "__main__":
    unittest.main()
