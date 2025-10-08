# Chrome Tab Capture

Direct browser tab screenshots using Chrome DevTools Protocol integration.

## Setup Requirements

Chrome must be launched with remote debugging enabled:

```bash
# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222

# Or with custom user data directory
"C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --user-data-dir=temp-profile
```

## Basic Tab Capture

### List Available Tabs

```bash
# Get all Chrome tabs
curl "http://localhost:8080/api/chrome/tabs" | jq '.tabs[] | {title, url, id}'

# Response example:
# {
#   "title": "GitHub - Homepage",
#   "url": "https://github.com",
#   "id": "E4C5B42D-8F3A-4E2B-9A1C-3D7E8F9A0B1C"
# }
```

### Capture Specific Tab

```bash
# Capture by tab ID
curl "http://localhost:8080/api/chrome/capture?tabId=E4C5B42D-8F3A-4E2B-9A1C-3D7E8F9A0B1C" -o github.png

# Capture with options
curl "http://localhost:8080/api/chrome/capture?tabId=E4C5B42D-8F3A-4E2B-9A1C-3D7E8F9A0B1C&format=jpeg&quality=90&fullPage=true" -o github_full.jpg
```

## Advanced Integration

### JavaScript Client Example

```javascript
class ChromeTabCapture {
    constructor(serverUrl = 'http://localhost:8080') {
        this.serverUrl = serverUrl;
    }

    async getTabs() {
        const response = await fetch(`${this.serverUrl}/api/chrome/tabs`);
        return response.json();
    }

    async captureTab(tabId, options = {}) {
        const params = new URLSearchParams({
            tabId,
            format: options.format || 'png',
            quality: options.quality || 90,
            fullPage: options.fullPage || false,
            ...options
        });

        const response = await fetch(
            `${this.serverUrl}/api/chrome/capture?${params}`
        );
        
        if (!response.ok) {
            throw new Error(`Capture failed: ${response.statusText}`);
        }

        return response.blob();
    }

    async executeScript(tabId, script) {
        const response = await fetch(`${this.serverUrl}/api/chrome/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ tabId, script })
        });

        return response.json();
    }

    async captureAllTabs(options = {}) {
        const tabs = await this.getTabs();
        const captures = [];

        for (const tab of tabs.tabs) {
            try {
                const blob = await this.captureTab(tab.id, options);
                captures.push({
                    tab: tab,
                    image: blob,
                    timestamp: new Date()
                });
            } catch (error) {
                console.error(`Failed to capture tab ${tab.title}:`, error);
            }
        }

        return captures;
    }
}

// Usage example
const chromeCapture = new ChromeTabCapture();

// Capture all open tabs
chromeCapture.captureAllTabs({ format: 'png' })
    .then(captures => {
        captures.forEach((capture, index) => {
            const url = URL.createObjectURL(capture.image);
            const link = document.createElement('a');
            link.href = url;
            link.download = `tab_${index}_${capture.tab.title.replace(/[^a-z0-9]/gi, '_')}.png`;
            link.click();
        });
    });
```

### Python Integration

```python
#!/usr/bin/env python3
import requests
import json
import time
from urllib.parse import quote

class ChromeTabCapture:
    def __init__(self, server_url="http://localhost:8080"):
        self.server_url = server_url
        self.session = requests.Session()
    
    def get_tabs(self):
        """Get all Chrome tabs"""
        response = self.session.get(f"{self.server_url}/api/chrome/tabs")
        response.raise_for_status()
        return response.json().get('tabs', [])
    
    def capture_tab(self, tab_id, filename=None, **options):
        """Capture a specific Chrome tab"""
        params = {'tabId': tab_id}
        params.update(options)
        
        response = self.session.get(
            f"{self.server_url}/api/chrome/capture",
            params=params
        )
        response.raise_for_status()
        
        if filename:
            with open(filename, 'wb') as f:
                f.write(response.content)
        
        return response.content
    
    def execute_script(self, tab_id, script):
        """Execute JavaScript in a Chrome tab"""
        payload = {
            'tabId': tab_id,
            'script': script
        }
        
        response = self.session.post(
            f"{self.server_url}/api/chrome/execute",
            json=payload
        )
        response.raise_for_status()
        return response.json()
    
    def capture_all_tabs(self, output_dir="./chrome_captures"):
        """Capture screenshots of all open tabs"""
        import os
        os.makedirs(output_dir, exist_ok=True)
        
        tabs = self.get_tabs()
        results = []
        
        for i, tab in enumerate(tabs):
            try:
                # Safe filename
                safe_title = "".join(c for c in tab['title'] if c.isalnum() or c in (' ', '-', '_')).rstrip()
                filename = f"tab_{i:02d}_{safe_title[:50]}.png"
                filepath = os.path.join(output_dir, filename)
                
                # Capture tab
                self.capture_tab(tab['id'], filepath, format='png')
                
                results.append({
                    'tab': tab,
                    'filename': filename,
                    'filepath': filepath,
                    'success': True
                })
                
                print(f"✅ Captured: {tab['title']}")
                
            except Exception as e:
                results.append({
                    'tab': tab,
                    'error': str(e),
                    'success': False
                })
                print(f"❌ Failed: {tab['title']} - {e}")
        
        return results
    
    def wait_for_page_load(self, tab_id, timeout=10):
        """Wait for page to finish loading"""
        script = """
        new Promise((resolve) => {
            if (document.readyState === 'complete') {
                resolve(true);
            } else {
                window.addEventListener('load', () => resolve(true));
            }
        })
        """
        
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                result = self.execute_script(tab_id, script)
                if result.get('result'):
                    return True
            except:
                pass
            time.sleep(0.5)
        
        return False
    
    def capture_workflow(self, urls, output_dir="./workflow_captures"):
        """Navigate to URLs and capture each page"""
        import os
        os.makedirs(output_dir, exist_ok=True)
        
        # Get first available tab
        tabs = self.get_tabs()
        if not tabs:
            raise Exception("No Chrome tabs available")
        
        tab_id = tabs[0]['id']
        results = []
        
        for i, url in enumerate(urls):
            try:
                print(f"📍 Navigating to: {url}")
                
                # Navigate to URL
                nav_script = f"window.location.href = '{url}'"
                self.execute_script(tab_id, nav_script)
                
                # Wait for page load
                print("  ⏳ Waiting for page load...")
                if self.wait_for_page_load(tab_id, timeout=15):
                    print("  ✅ Page loaded")
                else:
                    print("  ⚠️  Page load timeout")
                
                # Small delay for dynamic content
                time.sleep(2)
                
                # Capture screenshot
                filename = f"workflow_{i:02d}_{url.replace('://', '_').replace('/', '_')}.png"
                filepath = os.path.join(output_dir, filename)
                
                self.capture_tab(tab_id, filepath, format='png', fullPage=True)
                
                results.append({
                    'url': url,
                    'filename': filename,
                    'filepath': filepath,
                    'success': True
                })
                
                print(f"  📸 Captured: {filename}")
                
            except Exception as e:
                results.append({
                    'url': url,
                    'error': str(e),
                    'success': False
                })
                print(f"  ❌ Failed: {url} - {e}")
        
        return results

# Example usage
if __name__ == "__main__":
    capture = ChromeTabCapture()
    
    try:
        # Test basic functionality
        tabs = capture.get_tabs()
        print(f"Found {len(tabs)} Chrome tabs")
        
        # Capture all current tabs
        if tabs:
            print("\n📸 Capturing all tabs...")
            results = capture.capture_all_tabs()
            
            successful = sum(1 for r in results if r['success'])
            print(f"\n✨ Completed: {successful}/{len(results)} tabs captured")
        
        # Example workflow capture
        test_urls = [
            "https://github.com",
            "https://stackoverflow.com",
            "https://docs.python.org"
        ]
        
        print(f"\n🔄 Running workflow capture for {len(test_urls)} URLs...")
        workflow_results = capture.capture_workflow(test_urls)
        
        successful_workflow = sum(1 for r in workflow_results if r['success'])
        print(f"✨ Workflow completed: {successful_workflow}/{len(workflow_results)} pages captured")
        
    except Exception as e:
        print(f"❌ Error: {e}")
```

### Node.js Integration

```javascript
const puppeteer = require('puppeteer');
const axios = require('axios');

class ChromeTabCapture {
    constructor(serverUrl = 'http://localhost:8080') {
        this.serverUrl = serverUrl;
        this.browser = null;
    }

    async initialize() {
        // Connect to existing Chrome instance
        this.browser = await puppeteer.connect({
            browserURL: 'http://localhost:9222'
        });
    }

    async getTabs() {
        const response = await axios.get(`${this.serverUrl}/api/chrome/tabs`);
        return response.data.tabs;
    }

    async captureTab(tabId, options = {}) {
        const params = new URLSearchParams({
            tabId,
            format: options.format || 'png',
            quality: options.quality || 90,
            ...options
        });

        const response = await axios.get(
            `${this.serverUrl}/api/chrome/capture?${params}`,
            { responseType: 'arraybuffer' }
        );

        return Buffer.from(response.data);
    }

    async automatedCapture(scenarios) {
        const results = [];

        for (const scenario of scenarios) {
            try {
                console.log(`🎯 Running scenario: ${scenario.name}`);

                // Create new page or use existing tab
                const page = await this.browser.newPage();

                // Set viewport
                if (scenario.viewport) {
                    await page.setViewport(scenario.viewport);
                }

                // Navigate to URL
                await page.goto(scenario.url, { waitUntil: 'networkidle2' });

                // Execute custom actions
                if (scenario.actions) {
                    for (const action of scenario.actions) {
                        await this.executeAction(page, action);
                    }
                }

                // Get tab ID from server
                const tabs = await this.getTabs();
                const currentTab = tabs.find(tab => tab.url === scenario.url);

                if (currentTab) {
                    // Capture using server
                    const imageBuffer = await this.captureTab(currentTab.id, {
                        format: scenario.format || 'png',
                        fullPage: scenario.fullPage || false
                    });

                    // Save to file
                    const fs = require('fs');
                    const filename = `scenario_${scenario.name.replace(/\s+/g, '_')}.${scenario.format || 'png'}`;
                    fs.writeFileSync(filename, imageBuffer);

                    results.push({
                        scenario: scenario.name,
                        filename,
                        success: true,
                        size: imageBuffer.length
                    });

                    console.log(`  ✅ Captured: ${filename} (${imageBuffer.length} bytes)`);
                } else {
                    throw new Error('Could not find tab in server');
                }

                await page.close();

            } catch (error) {
                results.push({
                    scenario: scenario.name,
                    success: false,
                    error: error.message
                });

                console.log(`  ❌ Failed: ${error.message}`);
            }
        }

        return results;
    }

    async executeAction(page, action) {
        switch (action.type) {
            case 'click':
                await page.click(action.selector);
                break;
            case 'type':
                await page.type(action.selector, action.text);
                break;
            case 'wait':
                await page.waitForTimeout(action.duration);
                break;
            case 'waitForSelector':
                await page.waitForSelector(action.selector);
                break;
            case 'scroll':
                await page.evaluate((y) => window.scrollTo(0, y), action.y);
                break;
            case 'screenshot':
                await page.screenshot({ path: action.filename });
                break;
        }
    }

    async close() {
        if (this.browser) {
            await this.browser.disconnect();
        }
    }
}

// Example usage
async function main() {
    const capture = new ChromeTabCapture();
    
    try {
        await capture.initialize();

        // Define test scenarios
        const scenarios = [
            {
                name: 'GitHub Homepage',
                url: 'https://github.com',
                viewport: { width: 1920, height: 1080 },
                format: 'png',
                fullPage: true
            },
            {
                name: 'Search Results',
                url: 'https://github.com/search?q=screenshot',
                actions: [
                    { type: 'waitForSelector', selector: '.repo-list' },
                    { type: 'wait', duration: 2000 }
                ],
                format: 'jpeg',
                fullPage: false
            },
            {
                name: 'User Profile',
                url: 'https://github.com/octocat',
                actions: [
                    { type: 'waitForSelector', selector: '.user-profile-nav' },
                    { type: 'scroll', y: 500 }
                ],
                format: 'png'
            }
        ];

        console.log('🚀 Starting automated capture scenarios...');
        const results = await capture.automatedCapture(scenarios);

        // Summary
        const successful = results.filter(r => r.success).length;
        console.log(`\n✨ Completed: ${successful}/${results.length} scenarios`);

        results.forEach(result => {
            const status = result.success ? '✅' : '❌';
            console.log(`  ${status} ${result.scenario}`);
        });

    } catch (error) {
        console.error('❌ Error:', error.message);
    } finally {
        await capture.close();
    }
}

// Run if called directly
if (require.main === module) {
    main().catch(console.error);
}

module.exports = ChromeTabCapture;
```

## Advanced Use Cases

### E2E Testing Integration

```bash
# Capture before and after test actions
curl "http://localhost:8080/api/chrome/capture?tabId=TAB_ID&format=png" -o before.png
# ... perform test actions ...
curl "http://localhost:8080/api/chrome/capture?tabId=TAB_ID&format=png" -o after.png

# Compare images for visual regression testing
```

### Monitoring Web Applications

```bash
# Scheduled captures for monitoring
*/5 * * * * curl "http://localhost:8080/api/chrome/capture?tabId=DASHBOARD_TAB&format=jpeg&quality=70" -o "/monitoring/dashboard_$(date +\%Y\%m\%d_\%H\%M).jpg"
```

### Full Page Screenshots

```bash
# Capture entire page content
curl "http://localhost:8080/api/chrome/capture?tabId=TAB_ID&fullPage=true&format=png" -o fullpage.png
```

## Troubleshooting

### Common Issues

1. **Chrome not found**: Ensure Chrome is running with `--remote-debugging-port=9222`
2. **Tab not accessible**: Check if tab is still open and accessible
3. **Capture timeout**: Some pages take time to load, increase timeout
4. **Permission errors**: Some pages block screenshots due to security policies

### Debug Commands

```bash
# Test Chrome connection
curl "http://localhost:9222/json" | jq '.[] | {title, url, id}'

# Server Chrome status
curl "http://localhost:8080/api/chrome/status"

# Detailed tab information
curl "http://localhost:8080/api/chrome/tabs" | jq '.tabs[] | {title, url, id, type}'
```

## Best Practices

1. **Tab Management**: Close unused tabs to improve performance
2. **Timing**: Wait for page load before capturing
3. **Error Handling**: Always handle tab closure and navigation errors
4. **Resource Management**: Don't keep too many tabs open simultaneously
5. **Security**: Be cautious with executing arbitrary JavaScript

## Next Steps

- Try [DevTools Integration](devtools-integration.md) for advanced browser control
- Explore [Visual Regression Testing](../testing/visual-regression.md) with Chrome tabs
- Learn about [Real-time Monitoring](../monitoring/real-time-monitoring.md) for web applications