/**
 * @xaligo/xaligo
 *
 * TypeScript wrapper for the xaligo WebAssembly module.
 *
 * The WASM binary (`wasm/xaligo.wasm`) and Go runtime glue (`wasm/wasm_exec.js`)
 * are bundled in this package and loaded at runtime.
 *
 * Usage (Node.js / VS Code extension):
 *
 * ```ts
 * import { loadXaligo } from '@xaligo/xaligo';
 *
 * const xaligo = await loadXaligo();
 *
 * // Render .xal DSL string → Excalidraw JSON string
 * const json = await xaligo.render(xalSrc);
 *
 * // Render with services.csv for legend / icon abbreviations
 * const json = await xaligo.renderWithServices(xalSrc, servicesCsvContent);
 * ```
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Result returned by the Go WASM functions. */
interface WasmResult {
  result?: string;
  error?: string;
}

/** Public API exposed after the WASM module is loaded. */
export interface XaligoWasm {
  /**
   * Convert a `.xal` DSL string into an Excalidraw JSON string.
   * Uses the embedded AWS service-catalog and SVG assets.
   *
   * @param xal - Contents of a `.xal` file
   * @returns Excalidraw JSON string
   * @throws Error if the DSL is invalid or rendering fails
   */
  render(xal: string): Promise<string>;

  /**
   * Convert a `.xal` DSL string into an Excalidraw JSON string,
   * applying icon abbreviation overrides and a service legend from
   * a `services.csv` string (same format as `--services` in the CLI).
   *
   * @param xal        - Contents of a `.xal` file
   * @param servicesCsv - Contents of a `services.csv` file
   * @returns Excalidraw JSON string
   * @throws Error if the DSL is invalid or rendering fails
   */
  renderWithServices(xal: string, servicesCsv: string): Promise<string>;
}

// ---------------------------------------------------------------------------
// Global augmentation for the functions registered by the Go WASM runtime
// ---------------------------------------------------------------------------

declare global {
  function xaligoRender(xal: string): WasmResult;
  function xaligoRenderWithServices(xal: string, servicesCsv: string): WasmResult;
}

// ---------------------------------------------------------------------------
// Loader
// ---------------------------------------------------------------------------

let _instance: XaligoWasm | undefined;

/**
 * Load and initialise the xaligo WASM module.
 *
 * The function is idempotent — subsequent calls return the same instance
 * without reloading the binary.
 *
 * @param wasmUrl - Optional URL / path to `xaligo.wasm`.
 *   When omitted the bundled `wasm/xaligo.wasm` next to this package is used.
 */
export async function loadXaligo(wasmUrl?: string): Promise<XaligoWasm> {
  if (_instance) return _instance;

  // 1. Load the Go WASM runtime glue (wasm_exec.js).
  //    In Node.js we `require` it; in a browser it must be loaded separately
  //    (e.g. via a <script> tag or bundler).
  if (typeof globalThis.Go === 'undefined') {
    // Node.js path — resolve wasm_exec.js bundled alongside this package.
    const { createRequire } = await import('module');
    const require = createRequire(import.meta.url);
    require('../wasm/wasm_exec.js');
  }

  // 2. Determine the WASM binary path.
  const resolvedUrl = wasmUrl ?? new URL('../wasm/xaligo.wasm', import.meta.url).toString();

  // 3. Instantiate the Go WASM module.
  const go = new globalThis.Go();

  let wasmBytes: ArrayBuffer;
  if (typeof globalThis.fetch !== 'undefined') {
    // Browser / worker
    const resp = await fetch(resolvedUrl);
    if (!resp.ok) throw new Error(`Failed to fetch xaligo.wasm: ${resp.statusText}`);
    wasmBytes = await resp.arrayBuffer();
  } else {
    // Node.js (VS Code extension host)
    const { readFile } = await import('fs/promises');
    const { fileURLToPath } = await import('url');
    const filePath = resolvedUrl.startsWith('file:')
      ? fileURLToPath(resolvedUrl)
      : resolvedUrl;
    wasmBytes = (await readFile(filePath)).buffer;
  }

  const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
  // `go.run` never resolves — it blocks until the WASM exits.
  // We deliberately do NOT await it; the promise is kept alive in the background.
  void go.run(result.instance);

  // 4. Wait until the global functions are registered.
  await waitFor(() => typeof globalThis.xaligoRender === 'function', 5000);

  _instance = {
    render(xal: string): Promise<string> {
      const res: WasmResult = globalThis.xaligoRender(xal);
      if (res.error) throw new Error(res.error);
      if (!res.result) throw new Error('xaligoRender returned empty result');
      return Promise.resolve(res.result);
    },

    renderWithServices(xal: string, servicesCsv: string): Promise<string> {
      const res: WasmResult = globalThis.xaligoRenderWithServices(xal, servicesCsv);
      if (res.error) throw new Error(res.error);
      if (!res.result) throw new Error('xaligoRenderWithServices returned empty result');
      return Promise.resolve(res.result);
    },
  };

  return _instance;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function waitFor(predicate: () => boolean, timeoutMs: number): Promise<void> {
  return new Promise((resolve, reject) => {
    if (predicate()) { resolve(); return; }
    const start = Date.now();
    const interval = setInterval(() => {
      if (predicate()) {
        clearInterval(interval);
        resolve();
      } else if (Date.now() - start > timeoutMs) {
        clearInterval(interval);
        reject(new Error('xaligo WASM did not initialise within the timeout'));
      }
    }, 10);
  });
}

// Extend globalThis so TypeScript knows about the `Go` constructor from wasm_exec.js.
declare global {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const Go: new () => any;
}
