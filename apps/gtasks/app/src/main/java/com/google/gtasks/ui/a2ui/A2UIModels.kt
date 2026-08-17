// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.google.gtasks.ui.a2ui

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

@Serializable
data class BoundValue(
    val literalString: String? = null,
    val literalBoolean: Boolean? = null,
    val literalNumber: Double? = null,
    val path: String? = null
) {
    // Helper to resolve value from local data model state
    fun resolve(dataModel: Map<String, JsonElement>): String {
        path?.let { p ->
            dataModel[p]?.let { element ->
                // Clean up quotes from JSON string representation
                val str = element.toString()
                return if (str.startsWith("\"") && str.endsWith("\"")) {
                    str.substring(1, str.length - 1)
                } else {
                    str
                }
            }
        }
        literalString?.let { return it }
        literalBoolean?.let { return it.toString() }
        literalNumber?.let { return it.toString() }
        return ""
    }

    fun resolveBoolean(dataModel: Map<String, JsonElement>): Boolean {
        path?.let { p ->
            dataModel[p]?.let { element ->
                return element.toString().toBoolean()
            }
        }
        literalBoolean?.let { return it }
        literalString?.let { return it.toBoolean() }
        return false
    }
}

@Serializable
data class ComponentWrapper(
    @SerialName("Card") val card: CardProps? = null,
    @SerialName("Column") val column: ColumnProps? = null,
    @SerialName("Row") val row: RowProps? = null,
    @SerialName("Text") val text: TextProps? = null,
    @SerialName("Button") val button: ButtonProps? = null,
    @SerialName("TextInput") val textInput: TextInputProps? = null,
    @SerialName("CheckBox") val checkBox: CheckBoxProps? = null,
    @SerialName("Image") val image: ImageProps? = null,
    @SerialName("MultipleChoice") val multipleChoice: MultipleChoiceProps? = null,
    @SerialName("Divider") val divider: JsonElement? = null
)

@Serializable
data class A2UIComponent(
    val id: String,
    val component: ComponentWrapper
)

@Serializable
data class ChildrenProps(
    val explicitList: List<String> = emptyList()
)

@Serializable
data class CardProps(
    val title: String? = null,
    val child: String? = null,
    val style: String? = null
)

@Serializable
data class ColumnProps(
    val alignment: String? = null,
    val distribution: String? = null,
    val gap: Int = 0,
    val children: ChildrenProps
)

@Serializable
data class RowProps(
    val alignment: String? = null,
    val distribution: String? = null,
    val gap: Int = 0,
    val children: ChildrenProps
)

@Serializable
data class TextProps(
    val text: BoundValue,
    val usageHint: String? = null,
    val style: String? = null
)

@Serializable
data class ButtonAction(
    val type: String,
    val name: String,
    val context: List<ButtonContext> = emptyList()
)

@Serializable
data class ButtonContext(
    val key: String,
    val value: BoundValue
)

@Serializable
data class ButtonProps(
    val child: String? = null,
    val primary: Boolean = false,
    val label: String? = null,
    val action: ButtonAction? = null
)

@Serializable
data class TextInputProps(
    val label: String,
    val required: Boolean = false,
    @SerialName("dataBindingPath") val dataBindingPath: String
)

@Serializable
data class CheckBoxProps(
    val label: BoundValue? = null,
    val value: BoundValue? = null
)

@Serializable
data class ImageProps(
    val url: BoundValue
)

@Serializable
data class MultipleChoiceOption(
    val label: BoundValue,
    val value: String
)

@Serializable
data class MultipleChoiceProps(
    val options: List<MultipleChoiceOption> = emptyList(),
    val selections: BoundValue? = null,
    val maxAllowedSelections: Int = 1
)

@Serializable
data class SurfaceUpdatePayload(
    @SerialName("surfaceId") val surfaceID: String,
    val components: List<A2UIComponent> = emptyList()
)

@Serializable
data class DataModelUpdatePayload(
    @SerialName("surfaceId") val surfaceID: String,
    val data: Map<String, JsonElement> = emptyMap()
)

@Serializable
data class BeginRenderingPayload(
    val root: String,
    @SerialName("surfaceId") val surfaceID: String,
    val styles: Map<String, String> = emptyMap()
)

@Serializable
data class A2UITransaction(
    val surfaceUpdate: SurfaceUpdatePayload,
    val dataModelUpdate: DataModelUpdatePayload? = null,
    val beginRendering: BeginRenderingPayload
)
