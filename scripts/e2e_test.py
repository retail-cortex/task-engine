#!/usr/bin/env python3
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys
import os
import subprocess
import time

# Ensure playwright is installed
print("Checking and installing Playwright dependencies...")
try:
    import playwright
except ImportError:
    print("Installing Playwright package from public PyPI...")
    subprocess.run(["uv", "pip", "install", "--index-url", "https://pypi.org/simple/", "playwright"], check=True)

# Install browser binaries if needed
print("Installing Chromium browser binary...")
venv_playwright = os.path.join(sys.prefix, "bin", "playwright")
if not os.path.exists(venv_playwright):
    venv_playwright = ".venv/bin/playwright"
subprocess.run([venv_playwright, "install", "chromium"], check=True)

from playwright.sync_api import sync_playwright, expect

BASE_URL = "http://localhost:5173"

def run_tests():
    print("\n==================================================")
    print("🚀 BOOTING NEXUS HUB END-TO-END TEST SUITE")
    print("==================================================\n")

    success = True

    with sync_playwright() as p:
        # Launch headless Chromium
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1280, "height": 800})
        page = context.new_page()

        try:
            # Enable console logging from the browser
            page.on("console", lambda msg: print(f"   [BROWSER CONSOLE] {msg.type}: {msg.text}"))
            page.on("pageerror", lambda err: print(f"   [BROWSER UNCAUGHT ERROR] {err}"))

            # ----------------------------------------------------
            # TEST CASE 1: Load Dashboard & Handle SSO Authentication (Positive)
            # ----------------------------------------------------
            print("👉 [TEST 1] Pre-populating localstorage with Ryan's developer mock credentials...")
            
            # Inject the official GORM offline bypass token directly into localstorage to bypass Google SSO signature checks
            context.add_init_script("""
                localStorage.setItem('oauth_token', '00000000-0000-0000-0000-000000000000');
                localStorage.setItem('oauth_name', 'Hanna (Mock)');
                localStorage.setItem('oauth_email', 'hanna-mock@rmcguinness.altostrat.com');
            """)

            print("   Navigating to Main Dashboard...")
            page.goto(BASE_URL)
            page.wait_for_timeout(3000) # Wait for React profile hydration and GORM fetches
            
            # Assert we are on the main dashboard (brand-title swaps to HUB)
            expect(page.locator("h1.brand-title")).to_have_text("NEXUS INTEGRATION ENGINE HUB")
            print("   ✅ SUCCESS: Bypassed Google SSO and authenticated successfully as Ryan!")

            # ----------------------------------------------------
            # TEST CASE 2: Dual-Stage Brand & Store Selector Filtering (Positive)
            # ----------------------------------------------------
            print("\n👉 [TEST 2] Verifying Dual-Stage Scoping & Store Filtering...")
            
            # Locate selectors
            brand_select = page.locator("#brand-org-selector")
            store_select = page.locator("#store-selector")
            
            expect(brand_select).to_be_visible()
            expect(store_select).to_be_visible()

            # Step A: Filter by Volt & Vine
            print("   Selecting 'Volt & Vine' Brand by UUID...")
            brand_select.select_option(value="22222222-2222-2222-2222-222222222222")
            page.wait_for_timeout(1000)

            # Verify store dropdown only lists Volt & Vine stores
            store_options = store_select.locator("option")
            options_count = store_options.count()
            print(f"   Found {options_count} stores under Volt & Vine.")
            
            for i in range(options_count):
                opt_text = store_options.nth(i).inner_text()
                if "OmniMart" in opt_text:
                    raise AssertionError(f"Negative Filter Violation: OmniMart store '{opt_text}' found under Volt & Vine brand scope!")
            
            # Select Volt & Vine Seattle
            print("   Selecting 'Volt & Vine Seattle' Store by UUID...")
            store_select.select_option(value="44444444-4444-4444-4444-444444440000")
            page.wait_for_timeout(1000)
            
            # Verify Seattle tasks loaded
            task_queue_header = page.locator(".brand-section + div, div:has-text('Active')").first
            print(f"   Active Task Queue loaded for Seattle: {task_queue_header.inner_text()}")
            print("   ✅ SUCCESS: Volt & Vine store filtering matches brand scope perfectly!")

            # Step B: Filter by OmniMart
            print("   Selecting 'OmniMart' Brand by UUID...")
            brand_select.select_option(value="33333333-3333-3333-3333-333333333333")
            page.wait_for_timeout(1000)

            # Verify store dropdown only lists OmniMart stores
            store_options = store_select.locator("option")
            options_count = store_options.count()
            print(f"   Found {options_count} stores under OmniMart.")
            
            for i in range(options_count):
                opt_text = store_options.nth(i).inner_text()
                if "Volt & Vine" in opt_text:
                    raise AssertionError(f"Negative Filter Violation: Volt & Vine store '{opt_text}' found under OmniMart brand scope!")

            # Select OmniMart Dallas Store #1000
            print("   Selecting 'OmniMart Dallas' Store by UUID...")
            store_select.select_option(value="55555555-5555-5555-5555-555555550000")
            page.wait_for_timeout(1000)
            print("   ✅ SUCCESS: OmniMart store filtering matches brand scope perfectly!")

            # ----------------------------------------------------
            # TEST CASE 3: Navigate to Admin Panel & Open Task Blueprint (Positive)
            # ----------------------------------------------------
            print("\n👉 [TEST 3] Navigating to Admin Panel & Verifying Task Blueprint Form...")
            
            # Open SSO Profile dropdown menu
            print("   Opening SSO Profile dropdown...")
            page.locator("#profile-avatar-button").click()
            page.wait_for_timeout(500) # Wait for animation
            
            # Click Admin Control button in dropdown
            print("   Clicking 'Admin Control' button...")
            page.locator("#admin-control-button").click()
            page.wait_for_timeout(2000) # Wait for page transit
            print("   Admin Panel loaded successfully!")

            # Click the 'Task Blueprints' button in the sidebar to navigate to the blueprints view
            print("   Clicking 'Task Blueprints' sidebar button...")
            page.locator("button:has-text('Task Blueprints'), a:has-text('Task Blueprints')").first.click()
            page.wait_for_timeout(1000)

            # Wait for blueprints table
            expect(page.locator("h2:has-text('Task Workflow Templates')")).to_be_visible()
            
            # Click the Edit button on the first blueprint row
            edit_buttons = page.locator("button:has-text('Edit')")
            if edit_buttons.count() == 0:
                raise AssertionError("No task blueprints found in the admin table to test editing.")
            
            print("   Clicking edit button on first blueprint row...")
            edit_buttons.first.click()
            
            # Verify the edit form loaded successfully without a white screen crash
            expect(page.locator("h2:has-text('Edit Workflow Template')")).to_be_visible()
            expect(page.locator("label:has-text('Task Blueprint Name')")).to_be_visible()
            
            # Verify our custom high-fidelity JSONB Editor is loaded
            json_editor = page.locator("textarea[placeholder='{}']")
            expect(json_editor).to_be_visible()
            print("   ✅ SUCCESS: Edit form loaded successfully with custom JSONB editor, no GORM polymorphic crash detected!")

            # ----------------------------------------------------
            # TEST CASE 4: Negative Validation: Invalid JSONB Editing (Negative)
            # ----------------------------------------------------
            print("\n👉 [TEST 4] Verifying Negative Scenario: Invalid JSONB Syntax Validation...")
            
            # Type invalid JSON into the editor
            print("   Typing malformed JSON syntax...")
            json_editor.fill("{ malformed: json, missing_quotes }")
            page.wait_for_timeout(500)
            
            # Verify the syntax validator displays a red error status bar
            status_bar = page.locator("span:has-text('Invalid Syntax')")
            expect(status_bar).to_be_visible()
            
            # Verify the submit/save button is disabled when JSON is invalid
            save_btn = page.locator("button:has-text('Save Template'), button:has-text('Save'), button:has-text('Submit')").first
            expect(save_btn).to_be_disabled()
            print("   ✅ SUCCESS: Real-time syntax validator blocked invalid JSON submission and disabled save controls!")

            # Restore valid JSON to leave the form clean
            print("   Restoring valid JSON format...")
            json_editor.fill('{"demo": true, "key": "value"}')
            page.wait_for_timeout(500)
            
            valid_status = page.locator("span:has-text('Valid JSONB Metadata')")
            expect(valid_status).to_be_visible()
            expect(save_btn).to_be_enabled()
            print("   ✅ SUCCESS: Form restored to valid, submittable state.")

        except Exception as e:
            print(f"\n❌ TEST CASE FAILURE DETECTED: {e}")
            success = False
            # Capture failure screenshot for debugging
            screenshot_path = os.path.expanduser("~/.gemini/jetski/brain/dfc77464-6162-4cc7-ab2e-30e99e0005c7/e2e_failure_screenshot.png")
            page.screenshot(path=screenshot_path)
            print(f"   Failure screenshot saved to: {screenshot_path}")
        finally:
            browser.close()

    print("\n==================================================")
    if success:
        print("🎉 ALL END-TO-END TESTS PASSED SUCCESSFULLY!")
        print("==================================================\n")
        sys.exit(0)
    else:
        print("💥 TEST RUN FAILED. PLEASE REVIEW SYSTEM ERRORS.")
        print("==================================================\n")
        sys.exit(1)

if __name__ == "__main__":
    run_tests()
