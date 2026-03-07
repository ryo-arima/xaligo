#!/usr/bin/env node
// bin.js — finds the platform-specific xaligo binary and executes it.
// This pattern follows the approach used by esbuild, biome, etc.

'use strict';

const { spawnSync } = require('child_process');
const { join } = require('path');
const os = require('os');

const SUPPORTED_PLATFORMS = {
  'darwin-arm64': '@xaligo/xaligo-darwin-arm64',
  'darwin-x64':   '@xaligo/xaligo-darwin-x64',
  'linux-arm64':  '@xaligo/xaligo-linux-arm64',
  'linux-x64':    '@xaligo/xaligo-linux-x64',
  'win32-x64':    '@xaligo/xaligo-win32-x64',
};

const platform = process.platform;
const arch = process.arch;
const key = `${platform}-${arch}`;
const pkgName = SUPPORTED_PLATFORMS[key];

if (!pkgName) {
  console.error(
    `[xaligo] Unsupported platform: ${key}\n` +
    `Supported: ${Object.keys(SUPPORTED_PLATFORMS).join(', ')}`
  );
  process.exit(1);
}

const binaryName = platform === 'win32' ? 'xaligo.exe' : 'xaligo';
let binaryPath;

try {
  binaryPath = require.resolve(`${pkgName}/bin/${binaryName}`);
} catch {
  console.error(
    `[xaligo] Could not find binary for ${key}.\n` +
    `Try reinstalling: npm install @xaligo/xaligo`
  );
  process.exit(1);
}

const { status, error } = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
});

if (error) {
  console.error(`[xaligo] Failed to spawn binary: ${error.message}`);
  process.exit(1);
}

process.exit(status ?? 0);
