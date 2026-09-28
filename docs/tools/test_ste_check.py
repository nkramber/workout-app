"""Tests of the STE checker (D-83).

Run: python3 -m unittest discover -s docs/tools -p 'test_*.py'
"""
import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
TOOL = os.path.join(HERE, "ste-check.py")
spec = importlib.util.spec_from_file_location("ste_check", TOOL)
ste = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ste)


def rules(text):
    """The rule ids of each finding of one Markdown text."""
    with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False, encoding="utf-8") as handle:
        handle.write(text)
    try:
        return [rule for _, rule, _ in ste.check(handle.name)]
    finally:
        os.unlink(handle.name)


class Passing(unittest.TestCase):
    def test_plain_text_passes(self):
        self.assertEqual(rules("# Title\n\nThe app plans a workout. The user logs each set.\n"), [])

    def test_can_must_and_will_pass(self):
        self.assertEqual(rules("The user can skip a set. The policy must check the load. The app will save it.\n"), [])

    def test_a_state_participle_passes(self):
        self.assertEqual(rules("The pull request is merged. The tree is clean.\n"), [])

    def test_fitness_nouns_pass(self):
        text = ("Training starts on day one. Running and rowing count as cardio.\n\n"
                "The plan holds a block of stretching. Programming comes from the policy.\n\n"
                "The app sorts by rating. Cycling is one of the modes. A rule of rounding sets each load.\n")
        self.assertEqual(rules(text), [])

    def test_code_tables_headings_and_front_matter_are_exempt(self):
        text = ("---\nname: x\ndescription: It should be skipped.\n---\n\n# Setting up\n\n"
                "| a | b |\n|---|---|\n| It should | be skipped; yes |\n\n"
                "```\nit's done; it should work\n```\n\nThe text uses `it should` in code.\n")
        self.assertEqual(rules(text), [])

    def test_a_list_of_names_is_exempt_from_the_length_rule(self):
        names = ", ".join(f"name{i}" for i in range(30))
        self.assertEqual(rules(f"The tools: {names}.\n"), [])

    def test_a_wrapped_sentence_counts_once(self):
        text = "The app plans a workout for\nthe user from the equipment list.\n"
        self.assertEqual(rules(text), [])


class Broken(unittest.TestCase):
    def test_a_modal_verb_fails(self):
        self.assertIn("3.2/3.4", rules("The user should rest.\n"))

    def test_a_perfect_tense_fails(self):
        self.assertIn("3.2/3.4", rules("The app has changed the plan.\n"))

    def test_passive_voice_fails(self):
        self.assertIn("3.6", rules("The plan is written by the model.\n"))

    def test_an_ing_form_after_a_preposition_fails(self):
        self.assertIn("3.5", rules("Start the app before writing the plan.\n"))

    def test_an_ing_form_at_the_start_fails(self):
        self.assertIn("3.5", rules("Writing the plan takes time.\n"))

    def test_a_contraction_fails(self):
        self.assertIn("4.2", rules("The app doesn't stop.\n"))

    def test_a_semicolon_fails(self):
        self.assertIn("8.1", rules("The app stops; the user waits.\n"))

    def test_a_long_sentence_fails(self):
        self.assertIn("6.3", rules(" ".join(["word"] * 26) + ".\n"))

    def test_a_long_numbered_step_fails(self):
        self.assertIn("5.1", rules("1. " + " ".join(["word"] * 21) + ".\n"))

    def test_a_long_paragraph_fails(self):
        self.assertIn("6.6", rules(" ".join(["The app runs."] * 7) + "\n"))


class Command(unittest.TestCase):
    def test_the_exit_code_reads_the_findings(self):
        with tempfile.TemporaryDirectory() as folder:
            good = os.path.join(folder, "good.md")
            bad = os.path.join(folder, "bad.md")
            with open(good, "w", encoding="utf-8") as handle:
                handle.write("The app runs.\n")
            with open(bad, "w", encoding="utf-8") as handle:
                handle.write("The app should run.\n")
            ok = subprocess.run([sys.executable, TOOL, good], capture_output=True, text=True)
            fail = subprocess.run([sys.executable, TOOL, good, bad], capture_output=True, text=True)
        self.assertEqual(ok.returncode, 0, ok.stdout)
        self.assertEqual(fail.returncode, 1, fail.stdout)
        self.assertIn("bad.md:1: rule 3.2/3.4", fail.stdout)


if __name__ == "__main__":
    unittest.main()
