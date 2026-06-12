package com.google.gtasks

import android.app.Application
import com.google.gtasks.di.AppContainer

class GTasksApplication : Application() {
    // Dependency Injection container instance
    lateinit var container: AppContainer

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
    }
}
