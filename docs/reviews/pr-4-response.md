# Pull request 4 - author response

Date: 2026-09-28. Review round 1 recorded head `81f74d6`, with the verdict "Changes required".

## P2-1: Source verification ignores recorded source and license URLs

**Result: full merit.**

The trigger reproduces at `81f74d6`. A manifest entry with a wrong `source_url` or a wrong `license_url` passed `sources.check` with no difference, because the check did not compare these two fields. The manifest check accepted a license URL on any host when its path held the license path, for example `https://example.org/licenses/by/2.0`. The committed manifest had no such error: `sources.py check` on the correction compares all fields, and it gives 0 differences.

**Correction.**

- `tools/spikes/recognition_set/sources.py`: `check` compares `source_url` and `license_url` with the source record too, for Commons and for Flickr through Openverse (D-97, D-102, D-105).
- `tools/spikes/recognition_set/manifest_check.py`: `license_url_ok` accepts only a page on `creativecommons.org` for the same license. A CC0 image needs a `/publicdomain/zero/` page or no URL. A public domain image needs a `/publicdomain/` page or no URL.

**Regression checks.**

- `SourcesTest.test_check_gives_each_changed_url`: a wrong Commons `source_url` and a wrong `license_url` each give a difference.
- `SourcesTest.test_check_gives_each_changed_flickr_url`: a correct Flickr entry gives no difference, and a wrong `source_url` and a wrong `license_url` each give a difference.
- `LicenseTest.test_the_license_url_must_be_on_creativecommons_org`: a license URL on another host, or on a host that only starts with `creativecommons.org`, fails.
- The three tests fail on the code of `81f74d6`: 3 failures. All 43 tests of the folder pass on the correction, and `make verify` passes.
- `python3 tools/spikes/recognition_set/sources.py check` on the correction: 0 differences with the sources.

## P2-2: License URL paths accept a prefix of the required license

Review round 2 recorded head `a1bf5db`, with the verdict "Changes required".

**Result: full merit.**

The trigger reproduces at `a1bf5db`. `license_url_ok("cc-by-sa-2.0", "https://creativecommons.org/licenses/by-sa/2.0evil/")` gave `True`, because the check used a prefix match. The public domain code also accepted a CC0 path. The committed manifest had no such URL.

**Correction.**

- `tools/spikes/recognition_set/manifest_check.py`: `license_url_ok` matches the full path:
  - the family, the version, and the jurisdiction port of the code,
  - then an optional deed page or legal code page, with an optional language.
- A CC0 image needs `/publicdomain/zero/1.0`. A public domain image needs `/publicdomain/mark/1.0`, or no URL. A URL with a query or a fragment fails.

**Regression checks.**

- `LicenseTest.test_the_license_url_must_be_on_creativecommons_org` adds 6 false cases:
  - the two URLs of the finding,
  - a public domain code with a CC0 path,
  - a port path for a code with no port, and a code with a port with a path with no port,
  - an extra path part.
- The same test adds 3 true cases: a port deed page, a legal code page, and a public domain mark page.
- The test fails on the code of `a1bf5db`. All 43 tests of the folder pass on the correction. The manifest check gives 0 errors, and `make verify` passes.
