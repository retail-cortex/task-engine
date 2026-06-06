---
name: a2ui-integration
description: Guidelines and schema constraints for transpiling and rendering A2UI components under version 0.8 stable in Gemini Enterprise.
---

# A2UI Gemini Enterprise Integration Skill

This skill contains constraints, design guidelines, and patterns for converting model card responses into A2UI v0.8 stable message payloads. Use this skill whenever you are editing the transpiler, agent prompts, or verifying client-facing UI elements.

---

## 1. Schema Constraints: v0.8 Stable ONLY

Gemini Enterprise only supports **A2UI v0.8 stable**. All returned layout elements and components must strictly conform to v0.8 specifications. 

> [!IMPORTANT]
> Never output v0.9 layout patterns. Doing so will trigger critical validation errors and render failure tracebacks in the chat window.

### Layout Validation Mappings:
- **Wrapper Required**: Column, Row, and Table children list containers MUST be wrapped inside the `explicitList` property:
  ```json
  "children": {
    "explicitList": ["child_id_1", "child_id_2"]
  }
}
  ```
- **Single-Child Cards**: The `Card` component only accepts a singular `child` string pointing to another component ID:
  ```json
  "Card": {
    "child": "column_wrapper_id"
  }
  ```
  If you have multiple children, you must wrap them in a `Column` or `Row` first, and point the card's `child` field to the container.

---

## 2. Interactive Dropdown Collapsing

To prevent long vertical stacks of buttons in the chat viewport (e.g. site/store selectors):
- If the card children contain **two or more** buttons with matching actions (such as `SET_STORE`), collapse them into a single selection:
  - Use a `MultipleChoice` component with `"maxAllowedSelections": 1`.
  - Save selections to a specific path variable (e.g. `{"path": "/store_switcher/selected"}`).
  - Add a single submit `Button` below that reads the value path in its action context:
    ```json
    "Button": {
      "action": {
        "name": "SET_STORE",
        "context": [{ "key": "siteID", "value": { "path": "/store_switcher/selected" } }]
      }
    }
    ```

---

## 3. Markdown Code Block Interception

Model outputs containing UI Card JSON must be cleanly intercepted and parsed by the event converter. 
- **Regex Robustness**: Use case-insensitive matching for code blocks (`re.IGNORECASE` | `re.DOTALL`).
- **Pattern**: `r'```(?:json)?\s*(.*?)\s*```'` to successfully parse uppercase ````JSON`, lowercase ````json`, or language-omitted ```` blocks.

---

## 4. Image Component Schema

Image components must bind the source URL using a `BoundValue` structure. Raw strings will fail validation.

### Valid Image Component
```json
{
  "id": "image_component_id",
  "component": {
    "Image": {
      "url": {
        "literalString": "https://service-url.run.app/image.png"
      }
    }
  }
}
```

---

## 5. Spatial Digital Twin Blueprints

Since client-side SVGs, coordinate mapping, and custom canvas overlays are unsupported in chat frames:
- **Server-Side Render**: Draw plans and dynamic beacons as a server-side SVG image.
- **Service Route**: Serve SVGs with MIME type `image/svg+xml` from the FastAPI app (e.g., `GET /api/v1/blueprint?layout=linear&x=55&y=112`).
- **Mapping**: Transpile any `"canvas"` definitions in the model output to a standard `Image` pointing to this dynamic endpoint.
