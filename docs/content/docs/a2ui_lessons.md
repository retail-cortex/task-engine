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

---

## 7. Typographic Schema Normalization

A2UI v0.8 stable strictly enforces a small set of typographic values for the `usageHint` (and `style`) properties of text components.

> [!IMPORTANT]
> The only valid A2UI v0.8 typographic values are: `"h1"`, `"h2"`, `"h3"`, `"body"`, and `"caption"`.

Custom or legacy layout strings such as `"primary"` or `"secondary"` will fail v0.8 schema validation. When this happens, Google Chat's card validator silently rejects the component, dropping all affected labels and descriptions from the rendering.
To prevent this, the Go and Python A2UI generators must pass all text styles through a normalization function:
- `"primary"` $\rightarrow$ `"body"`
- `"secondary"` $\rightarrow$ `"caption"`
- Any other invalid style $\rightarrow$ `"body"`

---

## 8. Hermetic Base64 Vector Inlining (Mixed Content Resolution)

In secure Gemini Enterprise environments (hosted over HTTPS), referencing image endpoints over local HTTP (e.g. `http://localhost:8081/api/v1/blueprint`) triggers browser **Mixed Content blocking**, resulting in broken floor plans.

To resolve this, the transpiler intercepts canvas nodes and dynamic floor plan Image elements, generates the SVG vectors dynamically on the server side, base64-encodes them, and returns a self-contained **Base64 Data URI** directly in the card payload:
```
"url": {
  "literalString": "data:image/svg+xml;base64,PHN2ZyB4bWxucz0..."
}
```
This enables the client to render the floor plan instantly and hermetically without initiating any external network requests, bypassing browser Mixed Content blocks entirely.

---

## 9. Flutter-Style Alignment & Spacing

To align and distribute children inside Column and Row containers, A2UI v0.8 supports both standard and Flutter-style properties. Populating both sets concurrently guarantees perfect layout rendering regardless of the client-side A2UI parser version:
*   **Cross-Axis Alignment (`crossAxisAlignment`)**: Controls alignment perpendicular to the container's main axis (maps from `alignment`: `start` $\rightarrow$ `"Start"`, `center`/`middle` $\rightarrow$ `"Center"`, `stretch` $\rightarrow$ `"Stretch"`, `end` $\rightarrow$ `"End"`).
*   **Main-Axis Alignment (`mainAxisAlignment`)**: Controls children spacing and distribution along the container's main axis (maps from `distribution`: `start` $\rightarrow$ `"Start"`, `center` $\rightarrow$ `"Center"`, `space-between`/`spaceBetween` $\rightarrow$ `"SpaceBetween"`, `space-around` $\rightarrow$ `"SpaceAround"`, `space-evenly` $\rightarrow$ `"SpaceEvenly"`).

---

## 10. Sizing Constraints & Button Stretching

A2UI v0.8 buttons do not possess explicit width or styling properties (like `width` or `fullWidth`). Sizing is controlled entirely by parent layout containers.

### The Column Stretching Hack
To force a button to expand to the full width of the card or cell (e.g. full cell width):
1. Place the button inside a `Column` container.
2. Set the Column's cross-axis alignment to `"stretch"` (`"crossAxisAlignment": "Stretch"`).
3. If the button is wrapped inside a `Row`, it will only occupy its content width. For single-button states (such as `"View Details"`, `"Start Step"`, or `"Complete Task"`), **omit the Row wrapper** and append the button directly as a child of the parent Column to trigger full-width stretching.

---

## 11. Direct-Return Tool Output (429 Rate-Limit Mitigation)

In A2UI card execution flows, a single user click (such as starting a step) can trigger multiple sequential LLM calls in rapid succession (e.g. Initial turn $\rightarrow$ `update_task_status` tool call $\rightarrow$ `get_task_details` tool call $\rightarrow$ final token response). In low-quota projects, this high requests-per-minute (RPM) rate triggers `429 RESOURCE_EXHAUSTED` rate-limit errors.

To mitigate this, implement a **direct-return card update** pattern:
1. **Tool Refactoring**: Optimize backend tool handlers (e.g. `claim_task`, `update_task_status`) to directly invoke and return the updated A2UI details card (`[A2UI_CARD_TASK_DETAILS_CACHED]`) in their tool response, rather than returning a generic success text string.
2. **LLM Prompt Instruction**: Instruct the model that these tools return the updated details card directly, and that it **must not** make a separate, sequential `get_task_details` call.
This cuts the LLM round-trip count per click from 3 down to 2, reduces card refresh latency by ~33%, and completely eliminates 429 quota exhaustion risks.

---

## 12. Resilient A2UI Error Card Fallbacks

Unhandled server exceptions (e.g. database timeouts, OIDC validation failures, or rate-limit drops) normally bubble up as a `500 Internal Server Error`, causing the Gemini Enterprise client to freeze or display a generic error.

To deliver a premium, resilient experience, wrap the A2A HTTP gateway execution in a global try-except middleware:
1. Intercept all request processing exceptions.
2. Extract the request's JSON-RPC `id` to preserve protocol compliance.
3. Synthesize a friendly A2UI Error Card displaying custom instructions based on the error type (429 Rate Limits, Database Connection resets, 401 Session Expirations) along with a collapsible typographic `caption` containing the full technical stack trace for debugging.
4. Return a `200 OK` response containing the error card parts, ensuring the client renders a graceful, themed error state.

