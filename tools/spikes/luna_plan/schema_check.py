"""A small JSON Schema validator for the Luna plan spike (D-100).

The standard library has no JSON Schema validator, and the spike adds no
dependency. This file reads the subset of keywords that the strict
structured outputs of OpenAI permit: type, properties, required,
additionalProperties, items, and enum. A keyword outside that subset is
an error, so the schema and the validator can not drift apart.
"""
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
SCHEMA_PATH = os.path.join(HERE, "plan_schema.json")
SCHEMA_VERSION = "luna-plan-schema-v1"
KNOWN = {"$comment", "type", "properties", "required", "additionalProperties", "items", "enum"}


def load_schema():
    with open(SCHEMA_PATH, encoding="utf-8") as handle:
        return json.load(handle)


def type_ok(value, name):
    if name == "object":
        return isinstance(value, dict)
    if name == "array":
        return isinstance(value, list)
    if name == "string":
        return isinstance(value, str)
    if name == "boolean":
        return isinstance(value, bool)
    if name == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if name == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    if name == "null":
        return value is None
    raise ValueError(f"unknown type {name}")


def validate(value, schema, path="$"):
    """Return a list of error strings. An empty list means a pass."""
    unknown = set(schema) - KNOWN
    if unknown:
        raise ValueError(f"{path}: the validator does not read {sorted(unknown)}")
    errors = []
    names = schema.get("type")
    if names is not None:
        names = names if isinstance(names, list) else [names]
        if not any(type_ok(value, name) for name in names):
            return [f"{path}: expected {'/'.join(names)}, got {type(value).__name__}"]
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{path}: {value!r} is not one of {schema['enum']}")
    if isinstance(value, dict):
        props = schema.get("properties", {})
        for key in schema.get("required", []):
            if key not in value:
                errors.append(f"{path}: missing {key}")
        if schema.get("additionalProperties") is False:
            for key in value:
                if key not in props:
                    errors.append(f"{path}: extra property {key}")
        for key, sub in props.items():
            if key in value:
                errors.extend(validate(value[key], sub, f"{path}.{key}"))
    if isinstance(value, list) and "items" in schema:
        for i, item in enumerate(value):
            errors.extend(validate(item, schema["items"], f"{path}[{i}]"))
    return errors


def check_text(text, schema=None):
    """Parse the raw output text of a provider and validate it.

    Return (plan, errors). The plan is None when the text is not JSON.
    """
    schema = schema or load_schema()
    try:
        plan = json.loads(text)
    except (TypeError, ValueError) as exc:
        return None, [f"$: not JSON ({exc.__class__.__name__})"]
    return plan, validate(plan, schema)
