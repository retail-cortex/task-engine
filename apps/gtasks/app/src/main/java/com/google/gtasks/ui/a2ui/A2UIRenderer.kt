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

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive

@Composable
fun A2UIRenderer(
    transaction: A2UITransaction,
    modifier: Modifier = Modifier,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    // 1. Reactive Data Model State Pool (Local state binding database)
    val dataModel = remember(transaction) {
        mutableStateMapOf<String, JsonElement>().apply {
            transaction.dataModelUpdate?.data?.let { putAll(it) }
        }
    }

    // 2. Flattened Component Indexes
    val componentsMap = remember(transaction) {
        transaction.surfaceUpdate.components.associateBy { it.id }
    }

    val rootId = transaction.beginRendering.root

    Box(modifier = modifier.fillMaxWidth()) {
        RenderComponent(
            componentId = rootId,
            componentsMap = componentsMap,
            dataModel = dataModel,
            onAction = onAction
        )
    }
}

@Composable
private fun RenderComponent(
    componentId: String,
    componentsMap: Map<String, A2UIComponent>,
    dataModel: MutableMap<String, JsonElement>,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    val a2uiComponent = componentsMap[componentId] ?: return
    val wrapper = a2uiComponent.component

    when {
        wrapper.card != null -> RenderCard(wrapper.card, componentsMap, dataModel, onAction)
        wrapper.column != null -> RenderColumn(wrapper.column, componentsMap, dataModel, onAction)
        wrapper.row != null -> RenderRow(wrapper.row, componentsMap, dataModel, onAction)
        wrapper.text != null -> RenderText(wrapper.text, dataModel)
        wrapper.button != null -> RenderButton(wrapper.button, dataModel, onAction)
        wrapper.textInput != null -> RenderTextInput(wrapper.textInput, dataModel)
        wrapper.checkBox != null -> RenderCheckBox(wrapper.checkBox, dataModel)
        wrapper.image != null -> RenderImage(wrapper.image, dataModel)
        wrapper.multipleChoice != null -> RenderMultipleChoice(wrapper.multipleChoice, dataModel)
        wrapper.divider != null -> RenderDivider()
    }
}

@Composable
private fun RenderCard(
    props: CardProps,
    componentsMap: Map<String, A2UIComponent>,
    dataModel: MutableMap<String, JsonElement>,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .glassmorphic(elevation = 8.dp),
        colors = CardDefaults.cardColors(containerColor = Color.Transparent)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            props.title?.let { t ->
                Text(
                    text = t,
                    style = MaterialTheme.typography.titleMedium,
                    color = GTasksTheme.colors.textPrimary,
                    modifier = Modifier.padding(bottom = 12.dp)
                )
            }
            props.child?.let { childId ->
                RenderComponent(childId, componentsMap, dataModel, onAction)
            }
        }
    }
}

@Composable
private fun RenderColumn(
    props: ColumnProps,
    componentsMap: Map<String, A2UIComponent>,
    dataModel: MutableMap<String, JsonElement>,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(props.gap.dp),
        horizontalAlignment = when (props.alignment) {
            "center" -> Alignment.CenterHorizontally
            "end" -> Alignment.End
            else -> Alignment.Start
        }
    ) {
        props.children.explicitList.forEach { childId ->
            RenderComponent(childId, componentsMap, dataModel, onAction)
        }
    }
}

@Composable
private fun RenderRow(
    props: RowProps,
    componentsMap: Map<String, A2UIComponent>,
    dataModel: MutableMap<String, JsonElement>,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(props.gap.dp),
        verticalAlignment = when (props.alignment) {
            "center" -> Alignment.CenterVertically
            "bottom" -> Alignment.Bottom
            else -> Alignment.Top
        }
    ) {
        props.children.explicitList.forEach { childId ->
            RenderComponent(childId, componentsMap, dataModel, onAction)
        }
    }
}

@Composable
private fun RenderText(
    props: TextProps,
    dataModel: Map<String, JsonElement>
) {
    val textValue = props.text.resolve(dataModel)
    
    Text(
        text = textValue,
        style = when (props.usageHint) {
            "header" -> MaterialTheme.typography.titleMedium
            "subHeader" -> MaterialTheme.typography.titleSmall
            "caption" -> MaterialTheme.typography.labelSmall
            else -> MaterialTheme.typography.bodyMedium
        },
        color = when (props.style) {
            "primary" -> GTasksTheme.colors.colorPrimary
            "secondary" -> GTasksTheme.colors.colorSecondary
            "accent" -> GTasksTheme.colors.colorAccent
            "critical" -> GTasksTheme.colors.colorCritical
            "muted" -> GTasksTheme.colors.textMuted
            else -> GTasksTheme.colors.textPrimary
        }
    )
}

@Composable
private fun RenderButton(
    props: ButtonProps,
    dataModel: Map<String, JsonElement>,
    onAction: (ButtonAction, Map<String, JsonElement>) -> Unit
) {
    val isPrimary = props.primary
    val label = props.label ?: ""

    val onClick: () -> Unit = {
        props.action?.let { action ->
            onAction(action, dataModel)
        }
    }

    if (isPrimary) {
        Button(
            onClick = onClick,
            colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorPrimary),
            shape = RoundedCornerShape(8.dp)
        ) {
            Text(text = label, color = Color.White, style = MaterialTheme.typography.labelLarge)
        }
    } else {
        OutlinedButton(
            onClick = onClick,
            colors = ButtonDefaults.outlinedButtonColors(contentColor = GTasksTheme.colors.colorPrimary),
            border = BorderStroke(1.dp, GTasksTheme.colors.colorPrimary),
            shape = RoundedCornerShape(8.dp)
        ) {
            Text(text = label, style = MaterialTheme.typography.labelLarge)
        }
    }
}

@Composable
private fun RenderTextInput(
    props: TextInputProps,
    dataModel: MutableMap<String, JsonElement>
) {
    var textVal by remember(props.dataBindingPath) {
        mutableStateOf(dataModel[props.dataBindingPath]?.toString()?.removeSurrounding("\"") ?: "")
    }

    OutlinedTextField(
        value = textVal,
        onValueChange = { newValue ->
            textVal = newValue
            dataModel[props.dataBindingPath] = JsonPrimitive(newValue)
        },
        label = { Text(text = props.label) },
        modifier = Modifier.fillMaxWidth(),
        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = GTasksTheme.colors.colorPrimary,
            unfocusedBorderColor = GTasksTheme.colors.borderMuted,
            focusedLabelColor = GTasksTheme.colors.colorPrimary,
            unfocusedLabelColor = GTasksTheme.colors.textSecondary,
            focusedTextColor = GTasksTheme.colors.textPrimary,
            unfocusedTextColor = GTasksTheme.colors.textPrimary
        )
    )
}

@Composable
private fun RenderCheckBox(
    props: CheckBoxProps,
    dataModel: MutableMap<String, JsonElement>
) {
    val labelText = props.label?.resolve(dataModel) ?: ""
    val bindingPath = props.value?.path ?: ""
    var checked by remember(bindingPath) {
        mutableStateOf(props.value?.resolveBoolean(dataModel) ?: false)
    }

    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .fillMaxWidth()
            .clickable {
                checked = !checked
                if (bindingPath.isNotEmpty()) {
                    dataModel[bindingPath] = JsonPrimitive(checked)
                }
            }
            .padding(vertical = 6.dp)
    ) {
        Checkbox(
            checked = checked,
            onCheckedChange = { isChecked ->
                checked = isChecked
                if (bindingPath.isNotEmpty()) {
                    dataModel[bindingPath] = JsonPrimitive(isChecked)
                }
            },
            colors = CheckboxDefaults.colors(checkedColor = GTasksTheme.colors.colorPrimary)
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = labelText,
            color = GTasksTheme.colors.textPrimary,
            style = MaterialTheme.typography.bodyMedium
        )
    }
}

@Composable
private fun RenderImage(
    props: ImageProps,
    dataModel: Map<String, JsonElement>
) {
    val url = props.url.resolve(dataModel)
    
    // Coil AsyncImage automatically loads SVG blueprint drawings as coil-svg is configured.
    AsyncImage(
        model = url,
        contentDescription = "Dynamic SVG Blueprint Map View",
        modifier = Modifier
            .fillMaxWidth()
            .height(220.dp)
            .clip(RoundedCornerShape(8.dp)),
        contentScale = ContentScale.Inside
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RenderMultipleChoice(
    props: MultipleChoiceProps,
    dataModel: MutableMap<String, JsonElement>
) {
    val bindingPath = props.selections?.path ?: ""
    var selectedVal by remember(bindingPath) {
        mutableStateOf(props.selections?.resolve(dataModel) ?: "")
    }

    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { expanded = !expanded },
        modifier = Modifier.fillMaxWidth()
    ) {
        val selectedLabel = props.options.find { it.value == selectedVal }?.label?.resolve(dataModel) ?: selectedVal
        
        OutlinedTextField(
            value = selectedLabel,
            onValueChange = {},
            readOnly = true,
            label = { Text("Select Storefront Site Context") },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
            modifier = Modifier
                .menuAnchor()
                .fillMaxWidth(),
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = GTasksTheme.colors.colorPrimary,
                unfocusedBorderColor = GTasksTheme.colors.borderMuted,
                focusedTextColor = GTasksTheme.colors.textPrimary,
                unfocusedTextColor = GTasksTheme.colors.textPrimary
            )
        )
        ExposedDropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false }
        ) {
            props.options.forEach { option ->
                val optionLabel = option.label.resolve(dataModel)
                DropdownMenuItem(
                    text = { Text(optionLabel) },
                    onClick = {
                        selectedVal = option.value
                        if (bindingPath.isNotEmpty()) {
                            dataModel[bindingPath] = JsonPrimitive(option.value)
                        }
                        expanded = false
                    }
                )
            }
        }
    }
}

@Composable
private fun RenderDivider() {
    HorizontalDivider(
        color = GTasksTheme.colors.borderMuted,
        thickness = 1.dp,
        modifier = Modifier.padding(vertical = 12.dp)
    )
}
