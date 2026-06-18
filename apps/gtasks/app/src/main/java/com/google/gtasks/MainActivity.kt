package com.google.gtasks

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.core.view.WindowCompat
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.google.gtasks.ui.screens.chat.ChatScreen
import com.google.gtasks.ui.screens.context.ContextScreen
import com.google.gtasks.ui.screens.detail.TaskDetailScreen
import com.google.gtasks.ui.screens.login.LoginScreen
import com.google.gtasks.ui.screens.tasks.TaskListScreen
import com.google.gtasks.ui.screens.translate.TranslationScreen
import com.google.gtasks.ui.theme.GTasksTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // 1. Draw full-bleed immersive screen behind status bar and navigation bar
        WindowCompat.setDecorFitsSystemWindows(window, false)

        val appContainer = (application as GTasksApplication).container

        setContent {
            GTasksTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = GTasksTheme.colors.bgMain
                ) {
                    GTasksNavigation(appContainer = appContainer)
                }
            }
        }
    }
}

@Composable
fun GTasksNavigation(appContainer: com.google.gtasks.di.AppContainer) {
    val navController = rememberNavController()
    
    // Read current auth states directly to determine initial route
    val isLoggedIn by appContainer.authRepository.isLoggedIn.collectAsState()
    val activeSiteId = appContainer.authRepository.activeSiteId
    
    // Reactively force navigation to login screen if the user becomes logged out (e.g. 401 Unauthorized)
    LaunchedEffect(isLoggedIn) {
        if (!isLoggedIn) {
            val currentRoute = navController.currentDestination?.route
            if (currentRoute != null && currentRoute != "login") {
                navController.navigate("login") {
                    popUpTo(0) { inclusive = true }
                }
            }
        }
    }

    val startDestination = when {
        !isLoggedIn -> "login"
        activeSiteId == null -> "store-context"
        else -> "tasks"
    }

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        // Route 1: Login Screen
        composable("login") {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate("store-context") {
                        popUpTo("login") { inclusive = true }
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // Route 2: Store Context Selector
        composable("store-context") {
            ContextScreen(
                onSiteSelected = {
                    navController.navigate("tasks") {
                        popUpTo("store-context") { inclusive = true }
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // Route 3: Tasks List (Active prioritized task executions)
        composable("tasks") {
            TaskListScreen(
                onTaskClick = { taskId ->
                    navController.navigate("detail/$taskId")
                },
                onChatClick = {
                    navController.navigate("chat")
                },
                onTranslateClick = {
                    navController.navigate("translate")
                },
                onLogout = {
                    appContainer.authRepository.logout()
                },
                onChangeSite = {
                    navController.navigate("store-context") {
                        popUpTo("tasks") { inclusive = true }
                    }
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // Route 4: Task Detail & Checklist
        composable(
            route = "detail/{taskId}",
            arguments = listOf(navArgument("taskId") { type = NavType.StringType })
        ) { backStackEntry ->
            val taskId = checkNotNull(backStackEntry.arguments?.getString("taskId"))
            TaskDetailScreen(
                taskId = taskId,
                onBackClick = {
                    navController.popBackStack()
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // Route 5: Conversational Coach Chat (Hanna Coach with native A2UI cards!)
        composable("chat") {
            ChatScreen(
                onBackClick = {
                    navController.popBackStack()
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // Route 6: Real-time Voice Translation Screen
        composable("translate") {
            TranslationScreen(
                onBackClick = {
                    navController.popBackStack()
                },
                modifier = Modifier.fillMaxSize()
            )
        }
    }
}
