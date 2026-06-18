package com.google.gtasks.ui.screens.detail

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.LockOpen
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Place
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.filled.SwapHoriz
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Map
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material.icons.filled.KeyboardArrowUp
import androidx.compose.ui.platform.LocalUriHandler
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.data.model.ChecklistItem
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.ui.screens.tasks.Colleague
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Person
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone
import java.util.Date

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskDetailScreen(
    taskId: String,
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: TaskDetailViewModel = viewModel(factory = TaskDetailViewModel.Factory)
) {
    val uiState by viewModel.uiState.collectAsState()
    val isCompletable by viewModel.isCompletable.collectAsState()
    val isSubmitting by viewModel.isSubmitting.collectAsState()

    var justificationText by remember { mutableStateOf("") }

    // Initial load
    LaunchedEffect(taskId) {
        viewModel.loadTaskDetails(taskId)
    }

    Scaffold(
        topBar = {
            // Top App Bar
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .glassmorphic(shape = RoundedCornerShape(0.dp), elevation = 4.dp)
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .statusBarsPadding()
                        .padding(horizontal = 12.dp, vertical = 14.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(onClick = onBackClick) {
                        Icon(
                            imageVector = Icons.Default.ArrowBack,
                            contentDescription = "Go Back",
                            tint = GTasksTheme.colors.textPrimary
                        )
                    }
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "TASK COMPLIANCE DETAILS",
                        style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                        color = GTasksTheme.colors.textPrimary
                    )
                }
            }
        },
        bottomBar = {
            // Bottom-pinned Complete Task Button
            val state = uiState
            if (state is TaskDetailUiState.Success) {
                val task = state.task
                if (task.status != "COMPLETED") {
                    val isCompletable = viewModel.isCompletable.value
                    val isSubmitting by viewModel.isSubmitting.collectAsState()
                    val colleagues by viewModel.colleagues.collectAsState()
                    val currentUserId = viewModel.currentUserId
                    var showTradeDialog by remember { mutableStateOf(false) }
                    
                    Box(
                        modifier = Modifier
                            .glassmorphic(shape = RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp), elevation = 8.dp)
                            .fillMaxWidth()
                            .navigationBarsPadding()
                            .padding(horizontal = 20.dp, vertical = 16.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            // Offer Trade Button (Only if assigned to current user!)
                            if (task.assigneeID == currentUserId) {
                                OutlinedButton(
                                    onClick = { showTradeDialog = true },
                                    enabled = !isSubmitting,
                                    colors = ButtonDefaults.outlinedButtonColors(contentColor = GTasksTheme.colors.colorPrimary),
                                    border = BorderStroke(1.dp, GTasksTheme.colors.colorPrimary.copy(alpha = 0.6f)),
                                    shape = RoundedCornerShape(12.dp),
                                    modifier = Modifier
                                        .weight(0.4f)
                                        .height(56.dp)
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.SwapHoriz,
                                        contentDescription = "Trade Task"
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text("Trade", fontWeight = FontWeight.Bold)
                                }
                            }
                            
                            // Complete Task Button
                            Button(
                                onClick = {
                                    viewModel.completeTask(taskId, onBackClick)
                                },
                                enabled = isCompletable && !isSubmitting && task.status == "IN_PROGRESS",
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = GTasksTheme.colors.colorAccent,
                                    disabledContainerColor = GTasksTheme.colors.textMuted.copy(alpha = 0.2f)
                                ),
                                shape = RoundedCornerShape(12.dp),
                                modifier = Modifier
                                    .weight(if (task.assigneeID == currentUserId) 0.6f else 1f)
                                    .height(56.dp)
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.Center
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Check,
                                        contentDescription = "Success check",
                                        tint = if (isCompletable && task.status == "IN_PROGRESS") Color.White else GTasksTheme.colors.textMuted
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = if (isSubmitting) "SYNCHRONIZING..." else "COMPLETE RUN",
                                        color = if (isCompletable && task.status == "IN_PROGRESS") Color.White else GTasksTheme.colors.textMuted,
                                        fontWeight = FontWeight.Bold,
                                        style = MaterialTheme.typography.labelLarge,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis
                                    )
                                }
                            }
                        }
                        
                        if (showTradeDialog) {
                            TradeProposeDialog(
                                taskName = task.task.name,
                                colleagues = colleagues,
                                onDismiss = { showTradeDialog = false },
                                onSelectColleague = { colleagueId ->
                                    viewModel.proposeTrade(task.id, colleagueId) {
                                        showTradeDialog = false
                                    }
                                }
                            )
                        }
                    }
                }
            }
        },
        containerColor = GTasksTheme.colors.bgMain,
        modifier = modifier
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .background(
                    Brush.radialGradient(
                        colors = listOf(
                            Color(0x1FEF4444), // Translucent Red (Critical/Warning Glow)
                            GTasksTheme.colors.bgMain
                        ),
                        radius = 1000f
                    )
                )
        ) {
            when (val state = uiState) {
                is TaskDetailUiState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = GTasksTheme.colors.colorPrimary)
                    }
                }
                is TaskDetailUiState.Success -> {
                    val task = state.task
                    val themeColors = GTasksTheme.colors
                    
                    var mapLocationToShow by remember { mutableStateOf<String?>(null) }
                    
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = 20.dp)
                            .verticalScroll(rememberScrollState())
                    ) {
                        Spacer(modifier = Modifier.height(20.dp))

                        // 0. Assignee Banner (RBAC Checked for Managers/Admins)
                        val activeRole = viewModel.activeRoleName
                        val isManager = activeRole == "ADMIN" || activeRole == "REGION_MANAGER" || activeRole == "SITE_MANAGER"
                        if (isManager && !task.assigneeID.isNullOrEmpty()) {
                            val assigneeName = if (task.assigneeID == viewModel.currentUserId) "You (Ryan)" else "Associate (Cashier)"
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(bottom = 12.dp)
                                    .background(themeColors.colorPrimary.copy(alpha = 0.15f), RoundedCornerShape(8.dp))
                                    .border(BorderStroke(1.dp, themeColors.colorPrimary.copy(alpha = 0.3f)), RoundedCornerShape(8.dp))
                                    .padding(horizontal = 14.dp, vertical = 8.dp)
                            ) {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Icon(Icons.Default.Person, contentDescription = null, tint = themeColors.colorPrimary, modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = "Assignee: $assigneeName | Status: ${task.status}",
                                        style = MaterialTheme.typography.bodySmall.copy(fontWeight = FontWeight.Bold),
                                        color = themeColors.textPrimary
                                    )
                                }
                            }
                        }

                        // 1. Task Summary Header Card
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .glassmorphic(shape = RoundedCornerShape(20.dp))
                        ) {
                            Column(modifier = Modifier.padding(20.dp)) {
                                val priorityLabel = when (task.priority) {
                                    1 -> "HIGH PRIORITY"
                                    2 -> "MEDIUM PRIORITY"
                                    else -> "LOW PRIORITY"
                                }
                                val priorityColor = when (task.priority) {
                                    1 -> GTasksTheme.colors.colorCritical
                                    2 -> GTasksTheme.colors.colorPrimary
                                    else -> GTasksTheme.colors.colorAccent
                                }
                                
                                Text(
                                    text = priorityLabel,
                                    color = priorityColor,
                                    style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.Bold)
                                )

                                Text(
                                    text = task.task.name,
                                    style = MaterialTheme.typography.displayMedium.copy(fontSize = 24.sp, fontWeight = FontWeight.Black),
                                    color = GTasksTheme.colors.textPrimary,
                                    modifier = Modifier.padding(top = 8.dp)
                                )

                                Text(
                                    text = task.description ?: task.task.description ?: "No description provided.",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = GTasksTheme.colors.textSecondary,
                                    modifier = Modifier.padding(top = 8.dp)
                                )



                                // Expandable Standard Operating Procedures (SOPs) Section
                                if (task.task.sops.isNotEmpty()) {
                                    Spacer(modifier = Modifier.height(12.dp))
                                    var sopsExpanded by remember { mutableStateOf(false) }
                                    val uriHandler = LocalUriHandler.current
                                    
                                    Column(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .background(GTasksTheme.colors.bgInput.copy(alpha = 0.5f), RoundedCornerShape(12.dp))
                                            .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted), RoundedCornerShape(12.dp))
                                            .padding(12.dp)
                                    ) {
                                        Row(
                                            modifier = Modifier
                                                .fillMaxWidth()
                                                .clickable { sopsExpanded = !sopsExpanded },
                                            horizontalArrangement = Arrangement.SpaceBetween,
                                            verticalAlignment = Alignment.CenterVertically
                                        ) {
                                            Row(verticalAlignment = Alignment.CenterVertically) {
                                                Icon(
                                                    imageVector = Icons.Default.LockOpen,
                                                    contentDescription = null,
                                                    tint = GTasksTheme.colors.colorPrimary,
                                                    modifier = Modifier.size(18.dp)
                                                )
                                                Spacer(modifier = Modifier.width(8.dp))
                                                Text(
                                                    text = "Standard Operating Procedures (${task.task.sops.size})",
                                                    style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                                                    color = GTasksTheme.colors.textPrimary
                                                )
                                            }
                                            Icon(
                                                imageVector = if (sopsExpanded) Icons.Default.KeyboardArrowUp else Icons.Default.KeyboardArrowDown,
                                                contentDescription = if (sopsExpanded) "Collapse" else "Expand",
                                                tint = GTasksTheme.colors.textSecondary
                                            )
                                        }
                                        
                                        if (sopsExpanded) {
                                            Spacer(modifier = Modifier.height(10.dp))
                                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                                task.task.sops.forEach { sop ->
                                                     Row(
                                                         modifier = Modifier
                                                             .fillMaxWidth()
                                                             .background(GTasksTheme.colors.bgMain.copy(alpha = 0.4f), RoundedCornerShape(8.dp))
                                                             .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted.copy(alpha = 0.5f)), RoundedCornerShape(8.dp))
                                                             .clickable {
                                                                 sop.canonicalUrl?.let { url ->
                                                                     try {
                                                                         uriHandler.openUri(url)
                                                                     } catch (e: Exception) {
                                                                         // ignore
                                                                     }
                                                                 }
                                                             }
                                                             .padding(horizontal = 12.dp, vertical = 10.dp),
                                                         verticalAlignment = Alignment.CenterVertically,
                                                         horizontalArrangement = Arrangement.SpaceBetween
                                                     ) {
                                                         Text(
                                                             text = sop.title,
                                                             style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Medium),
                                                             color = GTasksTheme.colors.colorPrimary,
                                                             modifier = Modifier.weight(1f)
                                                         )
                                                         Icon(
                                                             imageVector = Icons.Default.ChevronRight,
                                                             contentDescription = "Open Link",
                                                             tint = GTasksTheme.colors.textMuted,
                                                             modifier = Modifier.size(18.dp)
                                                         )
                                                     }
                                                }
                                            }
                                        }
                                    }
                                }

                                if (mapLocationToShow != null) {
                                    StoreMapDialog(
                                        siteId = viewModel.siteId,
                                        siteName = viewModel.siteName,
                                        locationName = mapLocationToShow!!,
                                        onDismiss = { mapLocationToShow = null }
                                    )
                                }
                            }
                        }

                        // 1.5. Task-Level Active Execution Controller Card (Start / Pause / Resume)
                        val itemBorderColor = when (task.status) {
                            "IN_PROGRESS" -> themeColors.colorAccent.copy(alpha = 0.6f)
                            "PAUSED" -> themeColors.colorWarning.copy(alpha = 0.6f)
                            else -> themeColors.borderMuted
                        }
                        if (task.status != "COMPLETED") {
                            Spacer(modifier = Modifier.height(14.dp))
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .glassmorphic(shape = RoundedCornerShape(16.dp))
                                    .border(BorderStroke(1.dp, itemBorderColor), RoundedCornerShape(16.dp))
                                    .padding(16.dp)
                            ) {
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.SpaceBetween
                                ) {
                                    Column {
                                        Text(
                                            text = "TASK ACTIVE WORKFLOW STATUS",
                                            style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                            color = themeColors.textSecondary
                                        )
                                        Text(
                                            text = task.status,
                                            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
                                            color = when (task.status) {
                                                "IN_PROGRESS" -> themeColors.colorAccent
                                                "PAUSED" -> themeColors.colorWarning
                                                else -> themeColors.textPrimary
                                            }
                                        )
                                    }
                                    
                                    Button(
                                        onClick = {
                                            when (task.status) {
                                                "PENDING" -> viewModel.startTask(task.id)
                                                "IN_PROGRESS" -> viewModel.pauseTask(task.id)
                                                "PAUSED" -> viewModel.resumeTask(task.id)
                                            }
                                        },
                                        enabled = !isSubmitting,
                                        colors = ButtonDefaults.buttonColors(
                                            containerColor = when (task.status) {
                                                "IN_PROGRESS" -> themeColors.colorWarning
                                                else -> themeColors.colorPrimary
                                            }
                                        ),
                                        shape = RoundedCornerShape(8.dp)
                                    ) {
                                        val icon = when (task.status) {
                                            "IN_PROGRESS" -> Icons.Default.Pause
                                            else -> Icons.Default.PlayArrow
                                        }
                                        val label = when (task.status) {
                                            "PENDING" -> "START TASK"
                                            "IN_PROGRESS" -> "PAUSE WORK"
                                            "PAUSED" -> "RESUME WORK"
                                            else -> ""
                                        }
                                        Icon(icon, contentDescription = null, modifier = Modifier.size(16.dp))
                                        Spacer(modifier = Modifier.width(6.dp))
                                        Text(label, fontWeight = FontWeight.Bold, style = MaterialTheme.typography.labelSmall)
                                    }
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(24.dp))

                        // 2. Constraints / Asset Bypass Block
                        val isCashDrop = task.task.name.contains("Cash Drop", ignoreCase = true) || task.task.id.contains("d000fa44")
                        if (isCashDrop && task.status != "COMPLETED") {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .glassmorphic(shape = RoundedCornerShape(16.dp))
                                    .border(BorderStroke(1.dp, GTasksTheme.colors.colorWarning.copy(alpha = 0.3f)), RoundedCornerShape(16.dp))
                            ) {
                                Column(modifier = Modifier.padding(16.dp)) {
                                    Row(verticalAlignment = Alignment.CenterVertically) {
                                        Icon(
                                            imageVector = Icons.Default.Warning,
                                            contentDescription = "Warning",
                                            tint = GTasksTheme.colors.colorWarning
                                        )
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text(
                                            text = "SECURITY CONSTRAINT LOCK",
                                            style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                                            color = GTasksTheme.colors.colorWarning
                                        )
                                    }
                                    Text(
                                        text = "Register cash drawer limits exceeded. Requires vault drop verification. Submit supervisor override justification to bypass.",
                                        style = MaterialTheme.typography.bodySmall,
                                        color = GTasksTheme.colors.textSecondary,
                                        modifier = Modifier.padding(top = 8.dp, bottom = 16.dp)
                                    )

                                    OutlinedTextField(
                                        value = justificationText,
                                        onValueChange = { justificationText = it },
                                        label = { Text("Supervisor Justification Code") },
                                        modifier = Modifier.fillMaxWidth(),
                                        colors = OutlinedTextFieldDefaults.colors(
                                            focusedBorderColor = GTasksTheme.colors.colorWarning,
                                            unfocusedBorderColor = GTasksTheme.colors.borderMuted
                                        )
                                    )

                                    Spacer(modifier = Modifier.height(12.dp))

                                    Button(
                                        onClick = {
                                            viewModel.submitAssetOverride(
                                                executionId = taskId,
                                                assetId = "ca54ce11-0000-0000-0000-000000000004",
                                                justification = justificationText
                                            )
                                            justificationText = ""
                                        },
                                        enabled = justificationText.isNotEmpty() && !isSubmitting,
                                        colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorWarning),
                                        modifier = Modifier.align(Alignment.End)
                                    ) {
                                        Icon(Icons.Default.LockOpen, "Lock Icon", modifier = Modifier.size(16.dp))
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text("Authorize Override")
                                    }
                                }
                            }
                            Spacer(modifier = Modifier.height(24.dp))
                        }

                        // 3. Interactive Checklist Title
                        Text(
                            text = "COMPLIANCE CHECKLIST STEPS",
                            style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                            color = GTasksTheme.colors.textSecondary,
                            modifier = Modifier.padding(bottom = 12.dp)
                        )

                        // Checklist Steps
                        if (viewModel.checklist.isEmpty()) {
                            Text(
                                text = "No checklist steps specified.",
                                color = GTasksTheme.colors.textMuted,
                                style = MaterialTheme.typography.bodyMedium,
                                modifier = Modifier.padding(vertical = 8.dp)
                            )
                        } else {
                            viewModel.checklist.forEachIndexed { index, item ->
                                val isStepEnabled = task.status == "IN_PROGRESS" && 
                                    !isSubmitting && 
                                    (index == 0 || viewModel.checklist[index - 1].completed)
                                
                                ChecklistStepItem(
                                    item = item,
                                    onStartClick = { viewModel.startStep(taskId, item.step) },
                                    onPauseClick = { viewModel.pauseStep(taskId, item.step) },
                                    onResumeClick = { viewModel.resumeStep(taskId, item.step) },
                                    onCompleteClick = { viewModel.completeStep(taskId, item.step) },
                                    onMapClick = {
                                        val stepLocation = resolveLocationName(item.action)
                                        mapLocationToShow = if (stepLocation != "Store Floor") {
                                            stepLocation
                                        } else {
                                            resolveLocationName(task.task.name)
                                        }
                                    },
                                    enabled = isStepEnabled
                                )
                                Spacer(modifier = Modifier.height(10.dp))
                            }
                        }

                        Spacer(modifier = Modifier.height(100.dp)) // bottom padding
                    }
                }
                is TaskDetailUiState.Error -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Text(
                                text = state.message,
                                color = GTasksTheme.colors.colorCritical,
                                style = MaterialTheme.typography.bodyLarge,
                                textAlign = TextAlign.Center
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Button(
                                onClick = { viewModel.loadTaskDetails(taskId) },
                                colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorPrimary)
                            ) {
                                Text("Retry")
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ChecklistStepItem(
    item: ChecklistItem,
    onStartClick: () -> Unit,
    onPauseClick: () -> Unit,
    onResumeClick: () -> Unit,
    onCompleteClick: () -> Unit,
    onMapClick: () -> Unit,
    enabled: Boolean,
    modifier: Modifier = Modifier
) {
    val themeColors = GTasksTheme.colors
    val itemBorderColor = when (item.status) {
        "COMPLETED" -> themeColors.colorPrimary.copy(alpha = 0.4f)
        "IN_PROGRESS" -> themeColors.colorAccent.copy(alpha = 0.6f)
        "PAUSED" -> themeColors.colorWarning.copy(alpha = 0.6f)
        else -> themeColors.borderMuted
    }

    // Step Timer calculation
    var stepSeconds by remember { mutableStateOf(0) }
    LaunchedEffect(item.status, item.startedAt, item.totalPausedSeconds) {
        if (item.status == "IN_PROGRESS") {
            val startedTime = item.startedAt?.let { parseRfc3339(it) } ?: System.currentTimeMillis()
            while (true) {
                val now = System.currentTimeMillis()
                val elapsed = ((now - startedTime) / 1000).toInt()
                stepSeconds = maxOf(0, elapsed - item.totalPausedSeconds)
                kotlinx.coroutines.delay(1000)
            }
        } else if (item.status == "PAUSED" || item.status == "COMPLETED") {
            val startedTime = item.startedAt?.let { parseRfc3339(it) } ?: 0L
            val completedTime = item.completedAt?.let { parseRfc3339(it) } ?: System.currentTimeMillis()
            if (startedTime > 0L) {
                val elapsed = ((completedTime - startedTime) / 1000).toInt()
                stepSeconds = maxOf(0, elapsed - item.totalPausedSeconds)
            }
        }
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .glassmorphic(shape = RoundedCornerShape(12.dp), elevation = 2.dp)
            .border(BorderStroke(1.dp, itemBorderColor), RoundedCornerShape(12.dp))
            .padding(14.dp)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.fillMaxWidth()
        ) {
            // Left Action Controller Icon based on Step Status
            if (enabled) {
                when (item.status) {
                    "PENDING" -> {
                        IconButton(
                            onClick = onStartClick,
                            colors = IconButtonDefaults.iconButtonColors(contentColor = themeColors.colorPrimary),
                            modifier = Modifier.size(36.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.PlayArrow,
                                contentDescription = "Start Step"
                            )
                        }
                    }
                    "IN_PROGRESS" -> {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            IconButton(
                                onClick = onPauseClick,
                                colors = IconButtonDefaults.iconButtonColors(contentColor = themeColors.colorWarning),
                                modifier = Modifier.size(36.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Pause,
                                    contentDescription = "Pause Step"
                                )
                            }
                            Spacer(modifier = Modifier.width(4.dp))
                            IconButton(
                                onClick = onCompleteClick,
                                colors = IconButtonDefaults.iconButtonColors(contentColor = themeColors.colorAccent),
                                modifier = Modifier.size(36.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Check,
                                    contentDescription = "Complete Step"
                                )
                            }
                        }
                    }
                    "PAUSED" -> {
                        IconButton(
                            onClick = onResumeClick,
                            colors = IconButtonDefaults.iconButtonColors(contentColor = themeColors.colorPrimary),
                            modifier = Modifier.size(36.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.PlayArrow,
                                contentDescription = "Resume Step"
                            )
                        }
                    }
                    "COMPLETED" -> {
                        Icon(
                            imageVector = Icons.Default.Check,
                            contentDescription = "Completed Check",
                            tint = themeColors.colorPrimary,
                            modifier = Modifier.size(28.dp).padding(horizontal = 4.dp)
                        )
                    }
                }
            } else {
                // If not enabled (task not started, paused, or completed), just show status icon
                Icon(
                    imageVector = if (item.completed) Icons.Default.Check else Icons.Default.PlayArrow,
                    contentDescription = null,
                    tint = if (item.completed) themeColors.colorPrimary else themeColors.textMuted,
                    modifier = Modifier.size(24.dp)
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            // Step Action Text & Timer / Delta Badges
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = item.action,
                    color = if (item.completed) themeColors.textSecondary else themeColors.textPrimary,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = if (item.required) FontWeight.Bold else FontWeight.Normal
                    )
                )
                
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.padding(top = 4.dp)
                ) {
                    if (item.status == "IN_PROGRESS") {
                        Text(
                            text = "Active: ${formatDuration(stepSeconds)} / SLO: ${formatDuration(item.sloSeconds)}",
                            color = themeColors.colorAccent,
                            style = MaterialTheme.typography.labelSmall
                        )
                    } else if (item.status == "PAUSED") {
                        Text(
                            text = "Paused: ${formatDuration(stepSeconds)}",
                            color = themeColors.colorWarning,
                            style = MaterialTheme.typography.labelSmall
                        )
                    } else if (item.status == "COMPLETED") {
                        val delta = item.sloDeltaSeconds
                        if (delta != null) {
                            val isCompliant = delta <= 0
                            val label = if (isCompliant) "SLO compliant: -${formatDuration(-delta)}" else "SLO bottleneck: +${formatDuration(delta)}"
                            val badgeBg = if (isCompliant) Color(0x1F10B981) else Color(0x1FEF4444)
                            val badgeText = if (isCompliant) Color(0xFF10B981) else Color(0xFFEF4444)
                            
                            Box(
                                modifier = Modifier
                                    .background(badgeBg, RoundedCornerShape(4.dp))
                                    .padding(horizontal = 6.dp, vertical = 2.dp)
                            ) {
                                Text(
                                    text = label,
                                    color = badgeText,
                                    style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold)
                                )
                            }
                        }
                    } else {
                        // Pending
                        Text(
                            text = "SLO Target: ${formatDuration(item.sloSeconds)}",
                            color = themeColors.textMuted,
                            style = MaterialTheme.typography.labelSmall
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.width(8.dp))
            IconButton(
                onClick = onMapClick,
                colors = IconButtonDefaults.iconButtonColors(contentColor = themeColors.colorPrimary),
                modifier = Modifier.size(32.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Map,
                    contentDescription = "View Floorplan",
                    modifier = Modifier.size(20.dp)
                )
            }
        }
    }
}

// Utility Helpers for Duration Formatting and RFC3339 Parsing

private fun formatDuration(seconds: Int): String {
    val h = seconds / 3600
    val m = (seconds % 3600) / 60
    val s = seconds % 60
    return if (h > 0) String.format("%02d:%02d:%02d", h, m, s) else String.format("%02d:%02d", m, s)
}

private fun parseRfc3339(timestamp: String): Long {
    return try {
        val sdf = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US).apply {
            timeZone = TimeZone.getTimeZone("UTC")
        }
        sdf.parse(timestamp)?.time ?: System.currentTimeMillis()
    } catch (e: Exception) {
        System.currentTimeMillis()
    }
}

@Composable
private fun TradeProposeDialog(
    taskName: String,
    colleagues: List<Colleague>,
    onDismiss: () -> Unit,
    onSelectColleague: (String) -> Unit
) {
    Dialog(onDismissRequest = onDismiss) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .glassmorphic(shape = RoundedCornerShape(24.dp), elevation = 8.dp)
                .padding(24.dp)
        ) {
            Column {
                Text(
                    text = "Propose Task Trade",
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                    color = GTasksTheme.colors.textPrimary
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "Select a colleague to offer coverage for:\n$taskName",
                    style = MaterialTheme.typography.bodySmall,
                    color = GTasksTheme.colors.textSecondary
                )
                Spacer(modifier = Modifier.height(16.dp))

                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.heightIn(max = 240.dp)
                ) {
                    // Option 1: Open Pool Trade (Post to Available)
                    item {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(GTasksTheme.colors.colorPrimary.copy(alpha = 0.1f), RoundedCornerShape(8.dp))
                                .border(BorderStroke(1.dp, GTasksTheme.colors.colorPrimary.copy(alpha = 0.5f)), RoundedCornerShape(8.dp))
                                .clickable { onSelectColleague("") }
                                .padding(horizontal = 14.dp, vertical = 10.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column {
                                Text(
                                    text = "Post to Available Pool",
                                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                    color = GTasksTheme.colors.colorPrimary
                                )
                                Text(
                                    text = "Open to anyone at this location",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = GTasksTheme.colors.textSecondary
                                )
                            }
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = "Open Trade Sparkle",
                                tint = GTasksTheme.colors.colorPrimary
                            )
                        }
                    }

                    items(colleagues) { colleague ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(GTasksTheme.colors.bgInput, RoundedCornerShape(8.dp))
                                .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted), RoundedCornerShape(8.dp))
                                .clickable { onSelectColleague(colleague.id) }
                                .padding(horizontal = 14.dp, vertical = 10.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column {
                                Text(colleague.name, style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold), color = GTasksTheme.colors.textPrimary)
                                Text(colleague.roleName, style = MaterialTheme.typography.labelSmall, color = GTasksTheme.colors.textMuted)
                            }
                            Icon(androidx.compose.material.icons.Icons.Default.ChevronRight, contentDescription = null, tint = GTasksTheme.colors.textMuted)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))
                TextButton(onClick = onDismiss, modifier = Modifier.align(Alignment.End)) {
                    Text("Cancel", color = GTasksTheme.colors.textSecondary)
                }
            }
        }
    }
}
