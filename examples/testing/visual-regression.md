# Visual Regression Testing

Automated visual testing using screenshot comparison to detect UI changes and regressions.

## Overview

Visual regression testing captures screenshots of your application at different states and compares them to detect unintended changes. This is essential for maintaining visual consistency across deployments.

## Basic Setup

### Initialize Baseline Screenshots

```bash
# Create baseline directory
mkdir -p ./baselines/desktop ./baselines/mobile

# Capture desktop baselines
curl "http://localhost:8080/api/screenshot?method=title&target=MyApp&format=png" -o "./baselines/desktop/homepage.png"
curl "http://localhost:8080/api/screenshot?method=title&target=MyApp&format=png&region=0,0,1920,600" -o "./baselines/desktop/header.png"

# Capture mobile baselines (if using responsive design)
curl "http://localhost:8080/api/screenshot?method=title&target=MyApp&format=png&region=0,0,375,667" -o "./baselines/mobile/homepage.png"
```

### Compare Current vs Baseline

```bash
# Capture current state
curl "http://localhost:8080/api/screenshot?method=title&target=MyApp&format=png" -o "./current/homepage.png"

# Use image comparison tool (ImageMagick example)
compare -metric PSNR "./baselines/desktop/homepage.png" "./current/homepage.png" "./diff/homepage_diff.png"
```

## Automated Testing Scripts

### Python Visual Testing Framework

```python
#!/usr/bin/env python3
import os
import requests
import json
import time
from PIL import Image, ImageDraw
import numpy as np
from datetime import datetime
from pathlib import Path

class VisualRegressionTester:
    def __init__(self, server_url="http://localhost:8080", threshold=0.02):
        self.server_url = server_url
        self.threshold = threshold  # 2% difference threshold
        self.session = requests.Session()
        self.results = []
        
    def setup_directories(self, base_path="./visual-tests"):
        """Create necessary directories for testing"""
        self.base_path = Path(base_path)
        self.baseline_path = self.base_path / "baselines"
        self.current_path = self.base_path / "current"
        self.diff_path = self.base_path / "diffs"
        self.reports_path = self.base_path / "reports"
        
        for path in [self.baseline_path, self.current_path, self.diff_path, self.reports_path]:
            path.mkdir(parents=True, exist_ok=True)
    
    def capture_screenshot(self, test_config):
        """Capture a single screenshot based on test configuration"""
        params = {
            'method': test_config.get('method', 'title'),
            'target': test_config['target'],
            'format': 'png'
        }
        
        # Add optional parameters
        for key in ['region', 'quality', 'monitor']:
            if key in test_config:
                params[key] = test_config[key]
        
        response = self.session.get(f"{self.server_url}/api/screenshot", params=params)
        response.raise_for_status()
        return response.content
    
    def save_baseline(self, test_name, test_config):
        """Save a baseline screenshot for a test"""
        image_data = self.capture_screenshot(test_config)
        baseline_file = self.baseline_path / f"{test_name}.png"
        
        with open(baseline_file, 'wb') as f:
            f.write(image_data)
        
        print(f"✅ Baseline saved: {test_name}")
        return baseline_file
    
    def capture_current(self, test_name, test_config):
        """Capture current screenshot for comparison"""
        image_data = self.capture_screenshot(test_config)
        current_file = self.current_path / f"{test_name}.png"
        
        with open(current_file, 'wb') as f:
            f.write(image_data)
        
        return current_file
    
    def compare_images(self, baseline_path, current_path):
        """Compare two images and return difference percentage"""
        try:
            baseline = Image.open(baseline_path).convert('RGB')
            current = Image.open(current_path).convert('RGB')
            
            # Ensure images are same size
            if baseline.size != current.size:
                return {
                    'difference': 100.0,
                    'error': f"Size mismatch: {baseline.size} vs {current.size}",
                    'passed': False
                }
            
            # Convert to numpy arrays
            baseline_array = np.array(baseline)
            current_array = np.array(current)
            
            # Calculate pixel differences
            diff_array = np.abs(baseline_array.astype(float) - current_array.astype(float))
            
            # Calculate percentage difference
            total_pixels = baseline_array.shape[0] * baseline_array.shape[1]
            different_pixels = np.count_nonzero(diff_array.sum(axis=2) > 10)  # Threshold for considering pixels different
            difference_percentage = (different_pixels / total_pixels) * 100
            
            # Create difference image
            diff_image = Image.fromarray(diff_array.astype(np.uint8))
            
            # Highlight differences in red
            highlight_diff = Image.new('RGB', baseline.size, (0, 0, 0))
            highlight_pixels = np.where(diff_array.sum(axis=2) > 10)
            if len(highlight_pixels[0]) > 0:
                highlight_array = np.array(highlight_diff)
                highlight_array[highlight_pixels[0], highlight_pixels[1]] = [255, 0, 0]  # Red
                highlight_diff = Image.fromarray(highlight_array)
            
            return {
                'difference': difference_percentage,
                'passed': difference_percentage <= self.threshold,
                'diff_image': diff_image,
                'highlight_image': highlight_diff,
                'different_pixels': different_pixels,
                'total_pixels': total_pixels
            }
            
        except Exception as e:
            return {
                'difference': 100.0,
                'error': str(e),
                'passed': False
            }
    
    def run_test(self, test_name, test_config):
        """Run a single visual regression test"""
        print(f"🧪 Running test: {test_name}")
        
        baseline_file = self.baseline_path / f"{test_name}.png"
        
        # Create baseline if it doesn't exist
        if not baseline_file.exists():
            print(f"  📸 Creating baseline for {test_name}")
            self.save_baseline(test_name, test_config)
            return {
                'test': test_name,
                'status': 'baseline_created',
                'passed': True,
                'message': 'New baseline created'
            }
        
        # Capture current screenshot
        current_file = self.capture_current(test_name, test_config)
        
        # Compare images
        comparison = self.compare_images(baseline_file, current_file)
        
        # Save difference images
        if 'diff_image' in comparison:
            diff_file = self.diff_path / f"{test_name}_diff.png"
            highlight_file = self.diff_path / f"{test_name}_highlight.png"
            
            comparison['diff_image'].save(diff_file)
            comparison['highlight_image'].save(highlight_file)
            
            comparison['diff_file'] = str(diff_file)
            comparison['highlight_file'] = str(highlight_file)
        
        result = {
            'test': test_name,
            'config': test_config,
            'baseline_file': str(baseline_file),
            'current_file': str(current_file),
            'difference': comparison.get('difference', 0),
            'passed': comparison.get('passed', False),
            'status': 'passed' if comparison.get('passed', False) else 'failed',
            'threshold': self.threshold,
            'timestamp': datetime.now().isoformat(),
            'error': comparison.get('error')
        }
        
        # Add file paths if available
        for key in ['diff_file', 'highlight_file', 'different_pixels', 'total_pixels']:
            if key in comparison:
                result[key] = comparison[key]
        
        status_icon = "✅" if result['passed'] else "❌"
        print(f"  {status_icon} {test_name}: {result['difference']:.2f}% difference")
        
        if result.get('error'):
            print(f"    Error: {result['error']}")
        
        self.results.append(result)
        return result
    
    def run_test_suite(self, test_suite):
        """Run a complete test suite"""
        print(f"🚀 Running visual regression test suite: {test_suite['name']}")
        print(f"   Threshold: {self.threshold * 100}%")
        
        suite_results = {
            'name': test_suite['name'],
            'description': test_suite.get('description', ''),
            'timestamp': datetime.now().isoformat(),
            'threshold': self.threshold,
            'tests': [],
            'summary': {
                'total': len(test_suite['tests']),
                'passed': 0,
                'failed': 0,
                'baseline_created': 0
            }
        }
        
        for test_name, test_config in test_suite['tests'].items():
            result = self.run_test(test_name, test_config)
            suite_results['tests'].append(result)
            
            # Update summary
            if result['status'] == 'passed':
                suite_results['summary']['passed'] += 1
            elif result['status'] == 'failed':
                suite_results['summary']['failed'] += 1
            elif result['status'] == 'baseline_created':
                suite_results['summary']['baseline_created'] += 1
        
        # Generate report
        self.generate_html_report(suite_results)
        self.generate_json_report(suite_results)
        
        return suite_results
    
    def generate_html_report(self, suite_results):
        """Generate an HTML report"""
        html_content = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <title>Visual Regression Test Report - {suite_results['name']}</title>
            <style>
                body {{ font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }}
                .container {{ max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; }}
                .header {{ text-align: center; margin-bottom: 30px; }}
                .summary {{ background: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 30px; }}
                .test-result {{ margin-bottom: 20px; padding: 15px; border-radius: 5px; }}
                .test-result.passed {{ background: #d4edda; border-left: 4px solid #28a745; }}
                .test-result.failed {{ background: #f8d7da; border-left: 4px solid #dc3545; }}
                .test-result.baseline {{ background: #d1ecf1; border-left: 4px solid #17a2b8; }}
                .test-images {{ display: flex; gap: 10px; margin-top: 10px; }}
                .test-images img {{ max-width: 300px; border: 1px solid #ddd; }}
                .stats {{ display: flex; justify-content: space-around; }}
                .stat {{ text-align: center; }}
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>Visual Regression Test Report</h1>
                    <h2>{suite_results['name']}</h2>
                    <p>{suite_results.get('description', '')}</p>
                    <p>Generated: {suite_results['timestamp']}</p>
                </div>
                
                <div class="summary">
                    <h3>Summary</h3>
                    <div class="stats">
                        <div class="stat">
                            <h4>{suite_results['summary']['total']}</h4>
                            <p>Total Tests</p>
                        </div>
                        <div class="stat">
                            <h4 style="color: #28a745">{suite_results['summary']['passed']}</h4>
                            <p>Passed</p>
                        </div>
                        <div class="stat">
                            <h4 style="color: #dc3545">{suite_results['summary']['failed']}</h4>
                            <p>Failed</p>
                        </div>
                        <div class="stat">
                            <h4 style="color: #17a2b8">{suite_results['summary']['baseline_created']}</h4>
                            <p>Baselines Created</p>
                        </div>
                    </div>
                </div>
                
                <div class="tests">
        """
        
        for test in suite_results['tests']:
            status_class = test['status']
            if status_class == 'baseline_created':
                status_class = 'baseline'
            
            html_content += f"""
                    <div class="test-result {status_class}">
                        <h3>{test['test']}</h3>
                        <p><strong>Status:</strong> {test['status']}</p>
            """
            
            if test['status'] == 'failed' or test['status'] == 'passed':
                html_content += f"<p><strong>Difference:</strong> {test['difference']:.2f}% (threshold: {test['threshold']*100}%)</p>"
                
                if test.get('diff_file') and test.get('highlight_file'):
                    html_content += f"""
                        <div class="test-images">
                            <div>
                                <p>Baseline</p>
                                <img src="{os.path.relpath(test['baseline_file'], self.reports_path)}" alt="Baseline">
                            </div>
                            <div>
                                <p>Current</p>
                                <img src="{os.path.relpath(test['current_file'], self.reports_path)}" alt="Current">
                            </div>
                            <div>
                                <p>Differences Highlighted</p>
                                <img src="{os.path.relpath(test['highlight_file'], self.reports_path)}" alt="Differences">
                            </div>
                        </div>
                    """
            
            if test.get('error'):
                html_content += f"<p style='color: #dc3545'><strong>Error:</strong> {test['error']}</p>"
            
            html_content += "</div>"
        
        html_content += """
                </div>
            </div>
        </body>
        </html>
        """
        
        report_file = self.reports_path / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.html"
        with open(report_file, 'w') as f:
            f.write(html_content)
        
        print(f"📊 HTML report generated: {report_file}")
        return report_file
    
    def generate_json_report(self, suite_results):
        """Generate a JSON report"""
        report_file = self.reports_path / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        
        with open(report_file, 'w') as f:
            json.dump(suite_results, f, indent=2)
        
        print(f"📄 JSON report generated: {report_file}")
        return report_file

# Example test suite configuration
def create_sample_test_suite():
    return {
        'name': 'MyApp Visual Regression Tests',
        'description': 'Comprehensive visual testing for MyApp across different screen sizes and states',
        'tests': {
            'homepage_desktop': {
                'method': 'title',
                'target': 'MyApp - Home',
                'description': 'Desktop homepage view'
            },
            'homepage_mobile': {
                'method': 'title',
                'target': 'MyApp - Home',
                'region': '0,0,375,667',
                'description': 'Mobile homepage view'
            },
            'login_page': {
                'method': 'title',
                'target': 'MyApp - Login',
                'description': 'Login page'
            },
            'dashboard_main': {
                'method': 'title',
                'target': 'MyApp - Dashboard',
                'description': 'Main dashboard view'
            },
            'settings_dialog': {
                'method': 'class',
                'target': 'SettingsDialog',
                'description': 'Settings modal dialog'
            }
        }
    }

# Example usage
if __name__ == "__main__":
    # Initialize tester
    tester = VisualRegressionTester(threshold=0.05)  # 5% threshold
    tester.setup_directories()
    
    # Create and run test suite
    test_suite = create_sample_test_suite()
    results = tester.run_test_suite(test_suite)
    
    # Print summary
    print(f"\n✨ Test suite completed:")
    print(f"   Total: {results['summary']['total']}")
    print(f"   ✅ Passed: {results['summary']['passed']}")
    print(f"   ❌ Failed: {results['summary']['failed']}")
    print(f"   📸 Baselines Created: {results['summary']['baseline_created']}")
    
    # Exit with error code if tests failed
    import sys
    if results['summary']['failed'] > 0:
        sys.exit(1)
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/visual-regression.yml
name: Visual Regression Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  visual-tests:
    runs-on: windows-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
      
    - name: Setup Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install Pillow numpy requests
    
    - name: Start Screenshot Server
      run: |
        Start-Process -FilePath "./screenshot-server.exe" -ArgumentList "--port", "8080" -NoNewWindow
        Start-Sleep -Seconds 5
    
    - name: Start Application
      run: |
        Start-Process -FilePath "./MyApp.exe" -NoNewWindow
        Start-Sleep -Seconds 10
    
    - name: Run Visual Regression Tests
      run: python ./tests/visual_regression_test.py
    
    - name: Upload Test Results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: visual-test-results
        path: |
          ./visual-tests/reports/
          ./visual-tests/diffs/
    
    - name: Upload Screenshots on Failure
      uses: actions/upload-artifact@v3
      if: failure()
      with:
        name: failed-screenshots
        path: ./visual-tests/current/
    
    - name: Comment PR with Results
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const path = './visual-tests/reports';
          
          if (fs.existsSync(path)) {
            const reports = fs.readdirSync(path).filter(f => f.endsWith('.json'));
            if (reports.length > 0) {
              const report = JSON.parse(fs.readFileSync(`${path}/${reports[0]}`, 'utf8'));
              
              const comment = `## 📸 Visual Regression Test Results
              
**Summary:**
- Total Tests: ${report.summary.total}
- ✅ Passed: ${report.summary.passed}
- ❌ Failed: ${report.summary.failed}
- 📸 New Baselines: ${report.summary.baseline_created}

${report.summary.failed > 0 ? '⚠️ Some tests failed. Please review the visual differences.' : '🎉 All visual tests passed!'}

[View detailed report in artifacts]`;
              
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: comment
              });
            }
          }
```

### PowerShell CI Script

```powershell
# visual-regression-ci.ps1
param(
    [string]$ServerUrl = "http://localhost:8080",
    [string]$AppPath = "./MyApp.exe",
    [double]$Threshold = 0.05,
    [string]$OutputDir = "./visual-tests"
)

Write-Host "🚀 Visual Regression Testing Pipeline" -ForegroundColor Cyan

# Start screenshot server
Write-Host "Starting screenshot server..." -ForegroundColor Yellow
$serverProcess = Start-Process -FilePath "./screenshot-server.exe" -ArgumentList "--port", "8080" -PassThru -NoNewWindow
Start-Sleep -Seconds 5

# Test server connectivity
try {
    $healthCheck = Invoke-RestMethod -Uri "$ServerUrl/health" -Method Get -TimeoutSec 10
    Write-Host "✅ Screenshot server is running" -ForegroundColor Green
}
catch {
    Write-Host "❌ Failed to start screenshot server" -ForegroundColor Red
    exit 1
}

# Start application
Write-Host "Starting application..." -ForegroundColor Yellow
if (Test-Path $AppPath) {
    $appProcess = Start-Process -FilePath $AppPath -PassThru -NoNewWindow
    Start-Sleep -Seconds 10
    Write-Host "✅ Application started" -ForegroundColor Green
}
else {
    Write-Host "❌ Application not found: $AppPath" -ForegroundColor Red
    exit 1
}

# Run Python visual regression tests
Write-Host "Running visual regression tests..." -ForegroundColor Yellow
try {
    $testResult = python "./tests/visual_regression_test.py" 2>&1
    $exitCode = $LASTEXITCODE
    
    Write-Host $testResult
    
    if ($exitCode -eq 0) {
        Write-Host "✅ Visual regression tests passed" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Visual regression tests failed" -ForegroundColor Red
    }
}
catch {
    Write-Host "❌ Failed to run visual regression tests: $($_.Exception.Message)" -ForegroundColor Red
    $exitCode = 1
}

# Cleanup
Write-Host "Cleaning up processes..." -ForegroundColor Yellow
if ($serverProcess -and !$serverProcess.HasExited) {
    Stop-Process -Id $serverProcess.Id -Force
}
if ($appProcess -and !$appProcess.HasExited) {
    Stop-Process -Id $appProcess.Id -Force
}

# Generate CI summary
if (Test-Path "$OutputDir/reports") {
    $reportFiles = Get-ChildItem -Path "$OutputDir/reports" -Filter "*.json" | Sort-Object LastWriteTime -Descending
    if ($reportFiles.Count -gt 0) {
        $report = Get-Content -Path $reportFiles[0].FullName | ConvertFrom-Json
        
        Write-Host "`n📊 Test Summary:" -ForegroundColor Cyan
        Write-Host "   Total Tests: $($report.summary.total)" -ForegroundColor White
        Write-Host "   ✅ Passed: $($report.summary.passed)" -ForegroundColor Green
        Write-Host "   ❌ Failed: $($report.summary.failed)" -ForegroundColor Red
        Write-Host "   📸 Baselines Created: $($report.summary.baseline_created)" -ForegroundColor Blue
        
        if ($report.summary.failed -gt 0) {
            Write-Host "`n⚠️  Failed Tests:" -ForegroundColor Yellow
            $report.tests | Where-Object { $_.status -eq "failed" } | ForEach-Object {
                Write-Host "   • $($_.test): $($_.difference)% difference (threshold: $($_.threshold * 100)%)" -ForegroundColor Red
            }
        }
    }
}

exit $exitCode
```

## Best Practices

### Test Organization

1. **Group Related Tests**: Organize tests by feature or page
2. **Descriptive Names**: Use clear, descriptive test names
3. **Stable Baselines**: Only update baselines when intentional changes are made
4. **Environment Consistency**: Ensure consistent test environments

### Threshold Management

```python
# Different thresholds for different types of tests
THRESHOLDS = {
    'critical_ui': 0.01,    # 1% - Critical UI elements
    'layout': 0.02,         # 2% - Page layouts
    'content': 0.05,        # 5% - Dynamic content areas
    'animations': 0.10      # 10% - Areas with animations
}
```

### Handling Dynamic Content

```python
# Exclude dynamic regions from comparison
def mask_dynamic_regions(image, regions):
    """Mask dynamic regions before comparison"""
    draw = ImageDraw.Draw(image)
    for region in regions:
        draw.rectangle(region, fill=(128, 128, 128))  # Gray out region
    return image

# Usage
dynamic_regions = [(10, 50, 200, 100), (300, 200, 500, 250)]  # timestamp, ads, etc.
masked_image = mask_dynamic_regions(current_image, dynamic_regions)
```

## Next Steps

- Explore [CI/CD Integration](ci-integration.md) for automated testing pipelines
- Learn about [Real-time Monitoring](../monitoring/real-time-monitoring.md) for continuous visual validation
- Try [Chrome Tab Testing](../chrome/chrome-tabs.md) for web application visual testing