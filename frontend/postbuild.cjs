const fs = require('fs');
const path = require('path');

// Function to copy directory recursively
function copyRecursiveSync(src, dest) {
  const exists = fs.existsSync(src);
  const stats = exists && fs.statSync(src);
  const isDirectory = exists && stats.isDirectory();
  
  if (isDirectory) {
    if (!fs.existsSync(dest)) {
      fs.mkdirSync(dest, { recursive: true });
    }
    fs.readdirSync(src).forEach(childItemName => {
      copyRecursiveSync(
        path.join(src, childItemName),
        path.join(dest, childItemName)
      );
    });
  } else {
    fs.copyFileSync(src, dest);
  }
}

// Copy wailsjs to dist
const wailsjsSource = path.join(__dirname, 'wailsjs');
const wailsjsDest = path.join(__dirname, 'dist', 'wailsjs');

if (fs.existsSync(wailsjsSource)) {
  console.log('Copying wailsjs to dist...');
  copyRecursiveSync(wailsjsSource, wailsjsDest);
  console.log('wailsjs copied successfully!');
} else {
  console.log('Warning: wailsjs directory not found!');
}