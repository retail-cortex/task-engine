---
title: "A2UI Development Lessons"
weight: 50
---

# A2UI Integration Lessons Learned & Technical Reference

This document compiles the development history, architectural patterns, and integration guidelines established when integrating the **Agent Development Kit (ADK)** with the **A2UI Protocol** for Google Chat/Gemini Enterprise.

---

## 1. Version Constraint: v0.8 Stable

Google Gemini Enterprise chat clients natively enforce **A2UI v0.8 stable** schemas for rendering UI cards returned by agents. Adherence to v0.8 specifications is strictly required.

> [!WARNING]
> Do not use v0.9 or newer development draft definitions (such as direct arrays on layout lists). Doing so will result in protocol validation exceptions (`15 validation errors for SendMessageResponse`), causing the client to reject the payload and display a raw traceback or generic error block in the UI.

### Key differences in v0.8 vs v0.9:
- **Layout Children Wrappers**: v0.9 allows plain lists for children of columns/rows. v0.8 requires structural lists to be explicitly wrapped in an `explicitList` object.
- **Image URL Values**: v0.9 accepts raw strings for image URLs. v0.8 requires a `BoundValue` structure (using `literalString` or `path` pointers).
- **Single-Child Cards**: Card components in v0.8 only support a singular `child` string pointing to another component ID. Multi-children arrays are not permitted directly on a card.

---

## 2. Card Layout Rules

To output a structured card containing multiple elements (e.g., headers, text details, images, and action buttons):

1. Define a singular **Card** component as the root.
2. Direct the Card's `child` property to the component ID of a layout container component (e.g., a **Column**).
3. Place all children components inside the layout container's `explicitList` array:

```json
{
  "card_root": {
    "component": {
      "Card": {
        "child": "layout_column"
      }
    }
  },
  "layout_column": {
    "component": {
      "Column": {
        "children": {
          "explicitList": ["element_1", "element_2", "element_3"]
        }
      }
    }
  }
}
```

---

## 3. Interactive Dropdown Collapsing (Site Picker Pattern)

When listing multiple actions or items (such as storefront sites/locations for a context switcher), outputting a flat list of buttons creates a poor user experience and clutters the chat viewport. 

To resolve this, the transpiler dynamically collapses lists of two or more buttons targeting context switching (such as buttons with action `"SET_STORE"`) into a single-select picker:
1. **MultipleChoice Component**: Configured with `maxAllowedSelections: 1` to act as a single-select dropdown.
2. **Submit Button**: Placed below the dropdown, referencing the selection path (e.g., `{"path": "/store_switcher/selected"}`) to retrieve the chosen location ID and trigger the action.

```
┌──────────────────────────────────────┐
│  [Select Store Storefront Dropdown]  │
├──────────────────────────────────────┤
│  [Submit: Switch Store Button]       │
└──────────────────────────────────────┘
```

---

## 4. Markdown JSON Code Block Matching

When the agent model generates structured card JSON in response to a user request, it outputs standard markdown. The custom event converter interceptor on the FastAPI server extracts and parses this JSON structure before passing it to the A2A endpoint.

To prevent raw JSON block fallback in the client interface, match parameters are configured as:
- **Case-Insensitive Tags**: Matches must be case-insensitive to capture both uppercase ````JSON` (which Gemini Enterprise UI outputs by default for code blocks) and lowercase ````json`.
- **Optional Language Prefix**: The regex matches blocks with or without the language prefix (`r'```(?:json)?\s*(.*?)\s*```'`).
- **Whitespace Tolerance**: Newlines and surrounding text are stripped cleanly using `re.DOTALL` to ensure standard `json.loads` calls do not fail.

---

## 5. Image Component Specifications

The `Image` component in A2UI v0.8 requires its source URL to be passed as a bound literal structure.

### Correct v0.8 Image Schema
```json
{
  "id": "blueprint_map",
  "component": {
    "Image": {
      "url": {
        "literalString": "https://service.domain/api/v1/blueprint?layout=linear"
      }
    }
  }
}
```

> [!CAUTION]
> Specifying `"url": "https://..."` as a raw string is invalid and will trigger schema validation failures in the client.

---

## 6. Dynamic Server-Side Map Blueprints

A2UI v0.8 does not support dynamic client-side custom SVGs, inline canvas code, or interactive coordinate mapping. To overlay target coordinate beacons on digital twin floor blueprints, we leverage **server-side dynamic rendering**:

1. **Service Endpoint**: The task agent FastAPI app exposes a `GET /api/v1/blueprint` route.
2. **Request Parameters**: Accepts `layout` (e.g., `linear`, `boutique`, or `racetrack`) and coordinate parameters (`x`, `y`).
3. **SVG Generation**: The endpoint constructs the corresponding vector floor plan, overlays a pulsating beacon circle at the coordinate, and returns the response with MIME type `image/svg+xml`.
4. **Rendering**: The A2UI transpiler maps any `"canvas"` element in the model output to an `Image` component referencing this dynamic API URL, letting the client render the generated SVG floor plan natively.
