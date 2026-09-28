"""The prompt, the answer schema, and the providers of the recognition spike.

The prompt version is recognition-prompt-v1. The answer schema writes the
observed text and the evidence before the machine type, as section 6.2 of
docs/research/platform-cloud-and-ai.md recommends. Luna selects one type of
the catalog shortlist, or it abstains with `none_of_these` (a machine outside
the list) or `abstain` (the photo does not show enough).

Tests and CI use the fake provider only, so CI makes no paid call. The
OpenAI provider reads its model id from the role configuration (D-24), and it
sends each request through the call loop of the Luna plan spike, with the
same retries and the same cap gate.
"""
import base64
import hashlib
import json
import time

import photos  # noqa: F401  (puts the folders of the other spikes on the path)
import providers  # tools/spikes/luna_plan

PROMPT_VERSION = "recognition-prompt-v1"
ABSTAIN = ("none_of_these", "abstain")
QUALITY_FLAGS = ("blur", "low_light", "glare", "occlusion", "partial_view", "no_nameplate")

TEMPLATE = """You identify gym machines from one photo for a fitness app. The user confirms or corrects each answer, so an honest abstention is better than a guess.

Task:
- Select the one machine type of the catalog below that the photo shows. Use the type_id.
- Select "none_of_these" when the main machine of the photo is not in the catalog, for example a cable station, a rack, free weights, or outdoor equipment.
- Select "abstain" when the photo does not show enough of the machine to select one type.
- Look for the distinguishing detail of each type, such as plate horns or a weight stack, seated or standing, and the path of the movement.
- In observed_text, copy any text that you can read on the machine, such as a brand, a model, or a label. Write "" when you can read none.
- In evidence, give one or two short sentences about the parts of the machine that you see. Do not describe any person.
- In quality_flags, name each problem of the photo. Use an empty list when the photo is clear.
- In confidence, give the probability from 0 to 1 that your machine_type answer is correct.

Catalog:
{catalog}
"""


def instructions(catalog):
    rows = "\n".join(f"- {t['type_id']}: {t['name']}. {t['detail']}" for t in catalog["types"])
    return TEMPLATE.format(catalog=rows)


def prompt_hash(catalog):
    return hashlib.sha256(instructions(catalog).encode("utf-8")).hexdigest()[:16]


def answer_schema(catalog):
    """The strict JSON schema of one answer. The enum comes from the catalog."""
    type_ids = [t["type_id"] for t in catalog["types"]]
    return {
        "type": "object",
        "properties": {
            "observed_text": {"type": "string"},
            "evidence": {"type": "string"},
            "quality_flags": {"type": "array", "items": {"type": "string", "enum": list(QUALITY_FLAGS)}},
            "machine_type": {"type": "string", "enum": type_ids + list(ABSTAIN)},
            "confidence": {"type": "number"},
        },
        "required": ["observed_text", "evidence", "quality_flags", "machine_type", "confidence"],
        "additionalProperties": False,
    }


# The fake provider -------------------------------------------------------

# A fixed map of faults, so that a fake run shows each outcome of the report.
# The key is the photo id. "wrong_high" and "wrong_low" select another type.
DEFAULT_FAULTS = {
    "R002": "wrong_high", "R010": "wrong_low", "R020": "abstain", "R030": "none_of_these",
    "R040": "not_json", "R050": "refusal", "R060": "bad_confidence", "D003": "wrong_high",
    "D010": "abstain",
}


class FakeProvider:
    name = "fake"

    def __init__(self, catalog, faults=None):
        self.type_ids = [t["type_id"] for t in catalog["types"]]
        self.faults = DEFAULT_FAULTS if faults is None else faults

    def recognize(self, instructions_text, photo, image, schema, gate=None):
        gate = gate or providers.OpenGate()
        started = time.monotonic()
        if not gate.begin():
            return providers.Result("cap_skip", detail="the cap can not cover the attempt")
        fault = self.faults.get(photo["photo_id"])
        usage = providers.usage_dict(input_tokens=len(instructions_text) // 4 + 1200, reasoning=300)
        if fault == "refusal":
            usage["output_tokens"] = 20
            gate.end(usage)
            return providers.Result("refusal", detail="fake refusal", usage=usage,
                                    seconds=time.monotonic() - started)
        truth = photo["machine_type"]
        answer = {"observed_text": "", "evidence": "A fake answer.", "quality_flags": [],
                  "machine_type": truth, "confidence": 0.9}
        if fault in ("wrong_high", "wrong_low"):
            answer["machine_type"] = next(t for t in self.type_ids if t != truth)
            answer["confidence"] = 0.95 if fault == "wrong_high" else 0.4
        elif fault in ABSTAIN:
            answer["machine_type"], answer["confidence"] = fault, 0.3
        elif fault == "bad_confidence":
            answer["confidence"] = 7
        text = json.dumps(answer)
        if fault == "not_json":
            text = "The photo shows " + text
        usage["output_tokens"] = len(text) // 4 + usage["reasoning_tokens"]
        gate.end(usage)
        return providers.Result("completed", text=text, usage=usage, seconds=time.monotonic() - started)


# The OpenAI provider -----------------------------------------------------

class OpenAIProvider(providers.OpenAIProvider):
    """The OpenAI provider of the plan spike, with a request body for one photo."""

    def recognize_body(self, instructions_text, image, schema):
        url = "data:image/jpeg;base64," + base64.b64encode(image).decode("ascii")
        return {
            "model": self.role["model"],
            "reasoning": {"effort": self.role["reasoning_effort"]},
            "instructions": instructions_text,
            "input": [{"role": "user", "content": [
                {"type": "input_text", "text": "Identify the machine in this photo."},
                {"type": "input_image", "image_url": url, "detail": self.role["image_detail"]},
            ]}],
            "text": {"format": {"type": "json_schema", "name": "machine_answer", "schema": schema, "strict": True}},
            "max_output_tokens": self.role["max_output_tokens"],
            "store": False,
        }

    def recognize(self, instructions_text, photo, image, schema, gate=None):
        # The photo record holds the truth, and the request never reads it.
        return self.send(self.recognize_body(instructions_text, image, schema), gate)
