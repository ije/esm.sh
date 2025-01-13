# Changelog

## v136

* Add two builtin registries:
  * [jsr](https://jsr.io) - The open-source package registry for modern JavaScript and TypeScript, created by the Deno team. ([32cd2bd](https://github.com/esm-dev/esm.sh/commit/32cd2bd931f33118cbc96ee89583f20718c58fbf))
    ```js
    // example
    import { encodeBase64 } from "https://esm.sh/jsr/@std/encoding@1.0.0/base64";
    import { Hono } from "https://esm.sh/jsr/@hono/hono@4";
    ```
  * [pkg.pr.new](https://pkg.pr.new) - Continuous (Preview) Releases for your libraries, created by [StackBlitz Labs](https://github.com/stackblitz-labs) ([#904](https://github.com/esm-dev/esm.sh/pull/904), [#913](https://github.com/esm-dev/esm.sh/pull/913))
    ```js
    // Examples
    import { Bench } from "https://esm.sh/pr/tinylibs/tinybench/tinybench@a832a55";
    import { Bench } from "https://esm.sh/pr/tinybench@a832a55"; // --compact
    ```

* Respect semantic versioning for dependency resolving ([#875](https://github.com/esm-dev/esm.sh/pull/875))

  Previously, dependency resolution used fixed versions, which could cause duplication issues when packages were updated. Starting from v136, we now follow semantic versioning for dependencies.

  ```js
  // before
  "react-dom@19.0.0" import "/react@19.0.0/es2022/react.mjs";
  // after
  "react-dom@19.0.0" import "/react@^19.0.0?target=es2022";
  ```

* Add built-in npm package manager ([#948](https://github.com/esm-dev/esm.sh/pull/948))

  Implement a built-in npm package manager in Go to replace pnpm. This change reduces the CI test time from 12:15 to 4:30 (**~2.7x faster**) and eliminates the need for `nodejs` and `pnpm` dependencies.

* Split modules by analyzing the dependency tree during the build process ([#959](https://github.com/esm-dev/esm.sh/pull/959))

  To improve build performance and reduce network requests, esm.sh bundles sub-modules of a package by default. However, this can lead to duplicate code in the build. In v136, the server will split the modules by analyzing the dependency tree during the build process if the package's `exports` field is defined.

* Use [@pi0](https://github.com/pi0)'s [unenv](https://github.com/unjs/unenv) as the node runtime compatibility layer ([#914](https://github.com/esm-dev/esm.sh/pull/914))

  unenv provides a collection of Node.js and Web polyfills and mocking utilities with configurable presets for converting JavaScript code and libraries to be platform and runtime agnostic, working in any environment including Browsers, Workers, Node.js, Cloudflare Workers, Deno. unenv is also used by CloudFlare Workers: [blog](https://blog.cloudflare.com/more-npm-packages-on-cloudflare-workers-combining-polyfills-and-native-code/).

  You can also access these `unenv` node runtime modules directly via `/node/[node-builtin-module-name].mjs` path, for example:

  ```js
  import * from "https://esm.sh/node/fs.mjs";
  ```

* Add [npm-replacements](./server/npm_replacements/src/) that follows [e18e](https://e18e.dev)'s [module-replacements](https://github.com/es-tooling/module-replacements). ([#914](https://github.com/esm-dev/esm.sh/pull/914))

  The `npm-replacements` package replaces certain polyfill packages on NPM with native modern APIs during the build process. For example, the `object-assign` package is replaced with `Object.assign`:

  ```js
  import assign from "object-assign"; // replaced with "const assign = Object.assign"
  ```

* Remove build version prefix from the module path

  The build version prefix of esm.sh was introduced to avoid potential breaking changes caused by updates to the esm.sh server. However, it can lead to duplication issues when updating the server. In v136, we have removed the build version prefix from the module path, and the `?pin` query will be ignored. Paths with the build version prefix will continue to work as before, but the new default will be paths without the build version prefix.

  ```js
  // before
  "react@19.0.0" -> "https://esm.sh/v135/react@19.0.0/es2022/react.mjs";
  // after
  "react@19.0.0" -> "https://esm.sh/react@19.0.0/es2022/react.mjs";
  ```

* Deprecate the `build` API
  ```js
  import { build } from "https://esm.sh/build";
  build() // throws Error: The build API has been deprecated.
  ```

* Deprecate the Deno CLI script

  Deno now includes its own built-in package manager, which can be used to manage `jsr:` nad `npm:` imports in your deno applications. So we decided to deprecate the Deno CLI script.

  ```bash
  # use
  deno add npm:pract jsr:@std/encoding
  Add npm:preact@10.25.4
  Add jsr:@std/encoding@1.0.6

  # do not use
  deno run -A -r https://esm.sh init
  error: The deno CLI has been deprecated.
  ```

* Add `npmScopedRegistries` config

  The `npmScopedRegistries` configuration enables you to define the registry, token, user, and password for scoped npm packages. This allows you to import scoped packages using URLs like `https://your-esm-server.com/@scope_name/pkg@version`.

  ```json
  {
    "npmScopedRegistries": {
      "@scope_name": {
        "registry": "https://your-registry.com/",
        "token": "xxxxxx",
        "user": "",
        "password": ""
      }
    }
  }
  ```

* Experimental Features:
  * Build `.vue` and `.svelte` files on the fly ([#906](https://github.com/esm-dev/esm.sh/pull/906))
    ```js
    import "https://esm.sh/gh/phosphor-icons/vue@v2.2.0/src/icons/PhAirplay.vue?deps=vue@3.5.8"
    ```
  * https://esm.sh/x - ts/jsx/vue/svelte just works™️ in browser. ([#886](https://github.com/esm-dev/esm.sh/pull/886))

* Other Changes:
  * Upgrade esbuild to **0.24.2**
  * Use native [cjs-module-lexer@v1.0.7](https://github.com/esm-dev/cjs-module-lexer/releases/tag/v1.0.7)
  * Use target `es2022` for browsers by default ([#903](https://github.com/esm-dev/esm.sh/pull/903))
  * Replace `window` with `globalThis` to make Deno 2 happy ([#964](https://github.com/esm-dev/esm.sh/pull/964))
  * Use `.mjs` extenstion for sub-module build ([#917](https://github.com/esm-dev/esm.sh/pull/917))
  * dts-transformer: support `.d` and `d.cts` extension ([#987](https://github.com/esm-dev/esm.sh/pull/987))
  * Support `major.minor.patch+build` versioning ([#985](https://github.com/esm-dev/esm.sh/pull/985))
  * Support import attributes ([#976](https://github.com/esm-dev/esm.sh/pull/976))
  * config: Support **S3-compatible** storage ([#886](https://github.com/esm-dev/esm.sh/pull/886))
  * config: Add `corsAllowOrigins` config
  * config: Add `customLandingPage` config ([#928](https://github.com/esm-dev/esm.sh/pull/928))
  * config: Add `npmQueryCacheTTL` config ([#921](https://github.com/esm-dev/esm.sh/pull/921))

## v135

- Introduce https://esm.sh/run
- worker: Use `raw.esm.sh` hostname for ?raw option
- Add `?no-bundle` option
- Support `esm.sh` field in package.json
- Fix sub-module resolving (close [#754](https://github.com/esm-dev/esm.sh/issues/754), [#743](https://github.com/esm-dev/esm.sh/issues/743))
- Upgrade esbuild to **0.19.7**

## v134

- Add `transformOnly` option for build api
- Add `allowList` in config ([#745](https://github.com/esm-dev/esm.sh/pull/745) by @olekenneth)
- Improved Deno CLI ([#742](https://github.com/esm-dev/esm.sh/pull/742) by @Kyiro)
- Worker: fix dist version lookup
- Fix exported names from a dependency (close [#729](https://github.com/esm-dev/esm.sh/issues/729), [#750](https://github.com/esm-dev/esm.sh/issues/750))
- Fix: write `.npmrc` file if `NpmRegistry` is set (close [#737](https://github.com/esm-dev/esm.sh/pull/737)) ([#751](https://github.com/esm-dev/esm.sh/pull/751) by @edeustace)
- Upgrade esbuild to **0.19.5**

## v133

- Add `?raw` to support requests for raw package source files ([#731](https://github.com/esm-dev/esm.sh/pull/731) by @johnyanarella)
- Add global `setMaxListeners` to `node:events` polyfill ([#719](https://github.com/esm-dev/esm.sh/pull/719))
- cjs-lexer: resolving error now doesn't break build (close [#738](https://github.com/esm-dev/esm.sh/pull/738))
- Fix `cwd` method of `node:process` polyfill (close [#718](https://github.com/esm-dev/esm.sh/pull/718))
- Fix `applyConditions` function use `node` condition for browser (close [#732](https://github.com/esm-dev/esm.sh/pull/732))
- Fix `*.css.js` path (close [#728](https://github.com/esm-dev/esm.sh/pull/728))
- Fix some invalid _require_ imports (close [#724](https://github.com/esm-dev/esm.sh/pull/724))
- Fix relative path resolving of `browser` in package.json
- Upgrade esbuild to **0.19.4**

## v132

- Resolve node internal modules when `?external=*` set (close [#714](https://github.com/esm-dev/esm.sh/pull/714))
- Fix builds with `bigint` and `top-level-await` for all targets (close [#711](https://github.com/esm-dev/esm.sh/pull/711))
- Fix `node:process` ployfill module mssing the `hrtime` method
- Fix docker image missing `git` command
- esm-worker: add `varyUA` option for polyfill modules

## v131

- Add cache layer for the `/build` API
- Fix dts transformer resolver ignoring `*.mjs` url
- fix `?external` option ignoring sub-modules
- Use raw order of the `exports` in package.json (close [#705](https://github.com/esm-dev/esm.sh/pull/705))
- Redirect old build path (`.js`) to new build path (`.mjs`) (close [#703](https://github.com/esm-dev/esm.sh/pull/703))
- Upgrade esbuild to **0.19.2**

## v130

- esm-cjs-lexer: support minified UMD exports ([#689](https://github.com/esm-dev/esm.sh/pull/689))
- Support sub `.mjs` module (close [#691](https://github.com/esm-dev/esm.sh/pull/691))
- Fix `?bundle` mode ignores `node_process.js` (close [#694](https://github.com/esm-dev/esm.sh/pull/694))
- Upgrade `@types/react@18` to **18.2.15**
- Upgrade esbuild to **0.18.17**

## v129

- BREAKING: Remove `x-esm-deps` header (close [#683](https://github.com/esm-dev/esm.sh/issues/683))
- Sort `exports` of package.json when looping (close [#683](https://github.com/esm-dev/esm.sh/issues/683))
- Don't replace `typeof window` for deno target (close [#681](https://github.com/esm-dev/esm.sh/issues/681))
- Don't replace node global variable for `?target=node`
- Fix **dts** transformer (close [#670](https://github.com/esm-dev/esm.sh/issues/670))
- Fix depreacted message with `"`
- esm-worker: Fix cacheKey with `x-real-origin` header

## v128

- Add official Docker image: https://ghcr.io/esm-dev/esm.sh
- Fix missed `?external` of deps
- Fix duplicate `Access-Control-Expose-Headers` headers
- Fix dts transform for imports with both default and named imports ([#675](https://github.com/esm-dev/esm.sh/pull/675) by @hayes)
- Don't bundle dynamic imports
- Upgrade _stableBuild_ to **v128**

## v127

- Add `preload` imports
- Add `modern-normalize` to the `cssPackages`
- Fix subpath not be resovled with `?alias` (close [#671](https://github.com/esm-dev/esm.sh/issues/671))
- Fix dts transformer for "*.d" path imports (close [#660](https://github.com/esm-dev/esm.sh/issues/660))
- Fix source map mappings (close [#668](https://github.com/esm-dev/esm.sh/issues/668))
- CLI: Fix update failure caused by gh module ([#661](https://github.com/esm-dev/esm.sh/pull/661) by @lifegpc)
- Upgrade esbuild to **0.18.10**

## v126

- **breaking**: the `esm` tag function of build API now imorts module
  ```js
  import { esm } from "https://esm.sh/build";
  const mod = await esm`
    export const foo:string = "bar"
  `;
  console.log(mod.foo); // "bar"
  ```
- cjs-lexer: support _annotated_ exports (close [#659](https://github.com/esm-dev/esm.sh/issues/659))
- Add support for basic auth ([#657](https://github.com/esm-dev/esm.sh/pull/657) by @johnpangalos)

## v125

- Fix `node-fetch` import in cjs modules (close [#649](https://github.com/esm-dev/esm.sh/issues/649))
- Add `node:worker_threads` polyfill(fake) (close [#648](https://github.com/esm-dev/esm.sh/issues/648))
- Use `denonext` target for Deno >= 1.33.2 (close [#646](https://github.com/esm-dev/esm.sh/issues/646))
- Fix `.json.js` path (close [#645](https://github.com/esm-dev/esm.sh/issues/645))
- Fix cache missing content (close [#641](https://github.com/esm-dev/esm.sh/issues/641))
- Upgrade `deno/std` to **0.177.1**

## v124

- Fix the dts walker (close [#642](https://github.com/esm-dev/esm.sh/issues/642))

## v123

- Add `/server` endpoint for Deno to serve esm.sh locally
- Add scope to config ([#636](https://github.com/esm-dev/esm.sh/pull/636) by @johnpangalos)
- Fix `.d.ts` walker (close [#640](https://github.com/esm-dev/esm.sh/issues/640))
- Fix packages with `v` prefix in `version` (close [#639](https://github.com/esm-dev/esm.sh/issues/639))
- Fix `findFiles` function (close [#638](https://github.com/esm-dev/esm.sh/issues/638))

## v122

- Use stable imports order
- Support more asset extnames
- esm-worker: Use `X-Real-Origin` and `X-Esm-Worker-Version` headers
- Fix worker `CORS` issue (close [#631](https://github.com/esm-dev/esm.sh/issues/631))
- Fix sub-module resolving (close [#633](https://github.com/esm-dev/esm.sh/issues/633))
- Fix undefined content-type header (close [#635](https://github.com/esm-dev/esm.sh/issues/635))

## v121

- Use `browser` field for package main if possible
  ```json
  {
    "name": "pkg",
    "version": "1.0.0",
    "main": "./index.js",
    "browser": {
      "./index.js": "./browser.js"
    }
  }
  ```
- Fix redirects for `?css` and `GET /build`
- Fix `*.js.js` module path (close [#627](https://github.com/esm-dev/esm.sh/issues/627))
- Fix cjs imports (close [#629](https://github.com/esm-dev/esm.sh/issues/629), [#626](https://github.com/esm-dev/esm.sh/issues/626))
- Add `pako` to the `requireModeAllowList`

## v120

- build-api: Support types option
- Open-source the cloudflare worker
- Support `HEAD` method
- Fix bare path for css/custom build
- Fixing type only packages missing the `X-Typescript-Types` header
- Fix cjs-lexer `exports` resloving
- Use empty object instead of `null` for browser exclude (close [#613](https://github.com/esm-dev/esm.sh/issues/613))
- Add `zlib-sync` to nativeNodePackages (close [#621](https://github.com/esm-dev/esm.sh/issues/621))
- Redirect invalid `*.json` url

## v119

- Fix named import of cjs (close [#620](https://github.com/esm-dev/esm.sh/issues/620))
- Use `STABKE_VERSION` for dts build of `stableBuild`
- Upgrade esbuild to **0.17.18**

## v118

- feat: Publish system ([#607](https://github.com/esm-dev/esm.sh/pull/607))
- **esm-cjs-lexer**: Support `__export(require("..."))` pattern (close [#611](https://github.com/esm-dev/esm.sh/issues/611))
- Add `Auth` middleware
- Upgrade `stableBuild` to v118
- Remove **lit** from `stableBuild`
- Fix submodule types (close [#606](https://github.com/esm-dev/esm.sh/issues/606))
- Fix arch for darwin arm64 ([#617](https://github.com/esm-dev/esm.sh/pull/617) by @JLugagne)

## v117

- Fix Buffer polyfill for deno (close [#574](https://github.com/esm-dev/esm.sh/issues/574))
- Fix dts transformer with submodule (close [#599](https://github.com/esm-dev/esm.sh/issues/599))
- Fix importing `.json` as a module (close [#601](https://github.com/esm-dev/esm.sh/issues/601))
- Fix `.wasm` module importing (close [#602](https://github.com/esm-dev/esm.sh/issues/602))
- Fix path `/v100/PKG/TARGET/index.js`

## v116

- Support modules/assets from Github repo (close [#588](https://github.com/esm-dev/esm.sh/issues/588))
- Update `nativeNodePackages` (close [#591](https://github.com/esm-dev/esm.sh/issues/591))
- Fix dep import url of cjs module (close [#592](https://github.com/esm-dev/esm.sh/issues/592))
- Add support of resolving `typesVersions` (close [#593](https://github.com/esm-dev/esm.sh/issues/593))
- Fix `exports` glob condition resloving (close [#594](https://github.com/esm-dev/esm.sh/issues/594))
- Remove shebang (close [#596](https://github.com/esm-dev/esm.sh/issues/596))
- Fix missed build version of dts files (close [#589](https://github.com/esm-dev/esm.sh/issues/589))

## v115

- Return JavaScript modules for `?module` query with `wasm` files
- Fix types transformer (close [#581](https://github.com/esm-dev/esm.sh/issues/581))
- Fix incorrect named import of cjs modules (close [#583](https://github.com/esm-dev/esm.sh/issues/583))
- Fix sumodule path resolving (close [#584](https://github.com/esm-dev/esm.sh/issues/584))
- Upgrade `@types/node` to 18

## v114

- Add `?conditions` query as esbuild option
- Use **pnpm** to install packages instead of yarn (save the server disk space & improve the build performance)
- Serve static files on local ([#564](https://github.com/esm-dev/esm.sh/pull/564) @Justinidlerz)
- Support `.d.mts` extension (close [#580](https://github.com/esm-dev/esm.sh/issues/580))
- Fix cjs transpiling (close [#577](https://github.com/esm-dev/esm.sh/issues/577))
- Fix types bulid (close [#572](https://github.com/esm-dev/esm.sh/issues/572), [#576](https://github.com/esm-dev/esm.sh/issues/576))
- Fix invalid type URL if submodule is main entry ([#579](https://github.com/esm-dev/esm.sh/pull/579) @marvinhagemeister)
- Upgrade esbuild to 0.17.14

## v113

- `express` is working in Deno
- Fix lost non-mjs-extension module caused by v112 (close [#559](https://github.com/esm-dev/esm.sh/issues/559))
- Fix exports of `netmask` and `xml2js` ([#561](https://github.com/esm-dev/esm.sh/pull/561) @jcc10)
- Fix `default` import of deps for cjs (close [#565](https://github.com/esm-dev/esm.sh/issues/565), [#566](https://github.com/esm-dev/esm.sh/issues/566))

## v112

- Use `.mjs` extension for the package main module to resolve subpath conflicts
- Ignore `?exports` query when importing stable modules
- Fix npm naming regexp (close [#541](https://github.com/esm-dev/esm.sh/issues/541))
- Fix node buffer import for denonext target (closed [#556](https://github.com/esm-dev/esm.sh/issues/556))
- Fix tree shaking (close [#521](https://github.com/esm-dev/esm.sh/issues/521))
- Fix package nested conditions export ([#546](https://github.com/esm-dev/esm.sh/pull/546) by @Justinidlerz)
- Fix esm imports in cjs (close [#557](https://github.com/esm-dev/esm.sh/issues/557))
- Improve server performance ([#543](https://github.com/esm-dev/esm.sh/pull/543) by @Justinidlerz)
- Update requireModeAllowList (close [#540](https://github.com/esm-dev/esm.sh/issues/540), [#548](https://github.com/esm-dev/esm.sh/issues/548))

For Deno:
- Inject `XHR` polyfill for `axios`, `cross-fetch`, `whatwg-fetch` automatically
- CLI: Use user-specified indent size ([#551](https://github.com/esm-dev/esm.sh/pull/551) by @npg418)

## v111

- Print package `deprecated` message
- Remove source map url of worker
- Fix package CSS redirects with `target` option
- Fix build dead-loop for edge cases
- Fix CLI `update` command (close [#536](https://github.com/esm-dev/esm.sh/issues/536))

## v110

- Fix `Content-Type` header for dts files

## v109

- Ignore `?external` option for stable builds
- Fix `react/jsx-runtime` bundles `react` module
- Remove alias export resolving (close [#530](https://github.com/esm-dev/esm.sh/issues/530))

## v108

- Add `denonext` target to use [deno 1.31 node compatibility layer](https://deno.com/blog/v1.31#compatibility-layer-is-now-part-of-the-runtime)
- Redirect to css file for css packages
  ```
  https://esm.sh/normalize.css -> https://esm.sh/normalize.css/normalize.css
  ```
- Fix wasm packages can't get the wasm file.
  ```js
  import init, { transform } from "https://esm.sh/lightningcss-wasm";
  // before: you need to specify the wasm file path
  await init("https://esm.sh/lightningcss-wasm/lightningcss_node.wasm")
  // after: you don't need to specify it
  await init()
  ```
- Disable `bundle` mode for stable builds
- Fix alias export (close [#527](https://github.com/esm-dev/esm.sh/issues/527))
- Update references to reqOrigin to use cdnOrigin ([#529](https://github.com/esm-dev/esm.sh/pull/529) by [@jaredcwhite](https://github.com/jaredcwhite))

## v107

- Add `?cjs-export` query (close [#512](https://github.com/esm-dev/esm.sh/issues/512))<br>
  If you get an error like `...not provide an export named...`, that means esm.sh can not resolve CJS exports of the module correctly. You can add `?cjs-exports=foo,bar` query to specify the export names:
  ```javascript
  import { NinetyRing, NinetyRingWithBg } from "https://esm.sh/react-svg-spinners@0.3.1?cjs-exports=NinetyRing,NinetyRingWithBg"
  ```
- Update `requireModeAllowList` (close [#520](https://github.com/esm-dev/esm.sh/issues/520))
- **Remove** `?sourcemap` query, always generate source map as inline url.
- Default export all members from original module to prevent missing named exports members ([#522](https://github.com/esm-dev/esm.sh/pull/522))
- Only apply patch if types are missing in preact ([#523](https://github.com/esm-dev/esm.sh/pull/523))
- Upgrade `esbuild` to **0.17.10**.
- Upgrade `deno/std` to **0.177.0**

## v106

- Just fix fake module export names resolving (close [#510](https://github.com/esm-dev/esm.sh/issues/510))

## v105

- Check types which is not defined in `package.json`
- Fix empty module build (close [#483](https://github.com/esm-dev/esm.sh/issues/483))
- Fix exports field resolving (close [#503](https://github.com/esm-dev/esm.sh/issues/503))
- Fix deno cli script (close [#505](https://github.com/esm-dev/esm.sh/issues/505))
- Fix incorrect redirects (close [#508](https://github.com/esm-dev/esm.sh/issues/508))
- Fix invalid target with `HeadlessChrome/` UA (close [#509](https://github.com/esm-dev/esm.sh/issues/509))
- Upgrade `deno/std` to **0.175.0**

## v104

- Rewrite `FileSystem` interface of the storage.
- Fix submodule build with `exports` in package.json (close [#497](https://github.com/esm-dev/esm.sh/issues/497))
- Fix es5-ext weird `/#/` path (close [#502](https://github.com/esm-dev/esm.sh/issues/502))

## v103

- Add `inject` argument for worker factory
  ```js
  import workerFactory from "https://esm.sh/xxhash-wasm@1.0.2?worker";

  const workerInject = `
  self.onmessage = (e) => {
    // variable 'E' is the xxhash-wasm module default export
    E().then(hasher => {
      self.postMessage(hasher.h64ToString(e.data));
    })
  }
  `;

  const worker = workerFactory(workerInject);
  worker.onmessage = (e) => {
    console.log(e.data); // 502b0c5fc4a5704c
  };
  worker.postMessage("Hello");
  ```
- Respect `?external` arg in bundle mode (close [#498](https://github.com/esm-dev/esm.sh/issues/498))
- Add `require()` syntax support for **dts** transformer
- Fix import maps scope is not correct by the CLI script (close [#480](https://github.com/esm-dev/esm.sh/issues/480))
- Fix `basePath` doesn't take effect on redirects (close [#481](https://github.com/esm-dev/esm.sh/issues/481))
- Fix `X-TypeScript-Types` header not pined for stable builds
- Fix some bugs related to package path parsing ([#487](https://github.com/esm-dev/esm.sh/pull/487))
- Upgrade `esbuild` to **0.16.17**
- Upgrade `deno/std` to **0.173.0**

## v102

- Support `browser` field of **package.json** to improve compatibility with npm packages in browser. For example, the `webtorrent` package will use `memory-chunk-store` instead of `fs-chunk-store` and exclude built-in modules like `fs`, `net`, `os` and so on.
  ```json
  {
    "name": "webtorrent",
    "description": "Streaming torrent client",
    "version": "1.9.6",
    "browser": {
      "./lib/server.js": false,
      "./lib/conn-pool.js": false,
      "./lib/utp.js": false,
      "bittorrent-dht/client": false,
      "fs": false,
      "fs-chunk-store": "memory-chunk-store",
      "load-ip-set": false,
      "net": false,
      "os": false,
      "ut_pex": false
    },
  }
  ```
  (Close [#450](https://github.com/esm-dev/esm.sh/issues/450))

## v101

- Fix `?bundle` mode with illegal paths (close [#476](https://github.com/esm-dev/esm.sh/issues/476)).
- Fix `?worker` mode doesn't support CORS.

## v100

- Improve self-hosting configuration. check [HOSTING.md](./HOSTING.md) for more details.
- Support `browser` field when it's an es6 module (close [#381](https://github.com/esm-dev/esm.sh/issues/381)).
- Purge headers from unpkg.com to avoid repeated `Access-Control-Allow-Origin` header (close [#453](https://github.com/esm-dev/esm.sh/issues/453)).
- Fix content compression (close [#460](https://github.com/esm-dev/esm.sh/issues/460)).
- Fix alias export (close [#471](https://github.com/esm-dev/esm.sh/issues/471)).
- Fix cycle importing (close [#464](https://github.com/esm-dev/esm.sh/issues/464)).
- Fix scenarios where module/es2015 are shims (maps).
- Fix worker cors issue.
- Upgrade `esbuild` to **0.16.10**.
- Upgrade `deno/std` to **0.170.0**.

## v99

- Improve CDN cache performance, now you can get faster response time of `.d.ts`, `.wasm` and other static files.
- Remove `?deps` purge (close [#420](https://github.com/esm-dev/esm.sh/issues/420))
- Remove `?export` query of sub build task
- Upgrade `deno/std` to **0.165.0**.

## v98

- Add **tree-shaking** support for es modules
  ```js
  import { __await, __rest } from "https://esm.sh/tslib" // 7.3KB
  import { __await, __rest } from "https://esm.sh/tslib?exports=__await,__rest" // 489B
  ```
- Add `node-fetch` polyfill for browsers and deno
- Restart `ns` process when got "unreachable" error (close [#448](https://github.com/esm-dev/esm.sh/issues/448))
- Fix `exports` resolver (close [#422](https://github.com/esm-dev/esm.sh/issues/422))
- **cjs-lexer**: Update `swc` to latest

## v97

- Add `https://esm.sh/build-target` endpoint to return the build `target` of current browser/runtime by checking `User-Agent` header.
- Support `--npm-token` option to support private packages ([#435](https://github.com/esm-dev/esm.sh/issues/435)).
- Update `polyfills/node_process`: replace timeout with `queueMicrotask` ([#444](https://github.com/esm-dev/esm.sh/issues/444)).
- Upgrade `deno/std` to **0.162.0**.

## v96

- Update the fake node `fs` ployfill for browsers(add `createReadStream` and `createWriteStream` methods)
- Check package name (close [#424](https://github.com/esm-dev/esm.sh/issues/424))
- Fix some invalid types bulids
- Upgrade esbuild to **0.15.9**

## v95

- Fix `web-streams-ponyfill` build (close [#417](https://github.com/esm-dev/esm.sh/issues/417))
- Fix invalid `?deps` and `?alias` resolving
- Fix `solid-js/web` build for Deno
- Add `add react:preact/compat` pattern for the deno CLI

## v94

- Downgrade `deno/std` to **0.153.0**.

## v93

- Fix `@types/react` version (close [#331](https://github.com/esm-dev/esm.sh/issues/331))
- Fix cjs `__esModule` resolving (close [#410](https://github.com/esm-dev/esm.sh/issues/410))
- Fix `postcss-selector-parser` cjs exports (close [#411](https://github.com/esm-dev/esm.sh/issues/411))
- Fix `solid-js/web` of `deno` target
- Upgrade `deno/std` to **0.154.0**

## v92

- Add `stable` channel for UI libraries like react, to avoid multiple copies of runtime by cache
  ```
  https://esm.sh/v92/react@18.2.0/deno/react.js -> https://esm.sh/stable/react@18.2.0/deno/react.js
  ```
- Respect `external all` arg in types build
- Upgrade `deno/std` to **0.152.0**

## v91

- Improved Deno CLI Script:
  ```bash
  deno run -A https://esm.sh/v91 init
  ```
  After initializing, you can use the `deno task npm:[add/update/remove]` commands to manage the npm packages in the import maps.
  ```bash
  deno task npm:add react react-dom # add packages
  deno task npm:add react@17 react-dom@17 # add packages with specified version
  deno task npm:update react react-dom # upgrade packages
  deno task npm:update # update all packages
  deno task npm:remove react react-dom # remove packages
  ```
- Respect `imports` of package.json (close [#400](https://github.com/esm-dev/esm.sh/issues/400))
- Update `npmNaming` range (close [#401](https://github.com/esm-dev/esm.sh/issues/401))

## v90

- _Experimentally_ add Deno **CLI mode**, it will update the `import_map.json` file in the working directory:
  ```bash
  deno install -A -n esm -f https://esm.sh
  esm add react react-dom # add packages
  esm add react@17 react-dom@17 # add packages with specified version
  esm upgrade react react-dom # upgrade packages
  esm upgrade # upgrade all packages
  esm remove react react-dom # remove packages
  ```
  > Ensure to point the `import_map.json` in your `deno run` command or the `deno.json` file.
- Support `/v89/*some-package@version` external all pattern, do NOT use directly, use the CLI mode instead.
- Redirect urls with `/@types/` to the `.d.ts` file instead of build
- Improve node service stability
- Fix cjs `__exportStar` not used (close [#389](https://github.com/esm-dev/esm.sh/issues/389))
- Fix `resolve` package (close [#392](https://github.com/esm-dev/esm.sh/issues/392))
- Add workaround for `prisma` build
- Upgrade deno std to **0.151.0**

## v89

- support `?deno-std=$VER` to specify the [deno std](https://deno.land/std) version for deno node polyfills
- fix missed `__esModule` export

## v88

- Respect `exports.development` conditions in `package.json` (close [#375](https://github.com/esm-dev/esm.sh/issues/375))
- Fix `solid-js/web?target=deno` strip ssr functions
- Fix `@types/node` types transforming (close [#363](https://github.com/esm-dev/esm.sh/issues/363))
- Fix `?external` doesn't support `.dts` files (close [#374](h`ttps://github.com/esm-dev/esm.sh/issues/374))
- Fix invalid export names of `keycode`, `vscode-oniguruma` & `lru_map` (close [#362](https://github.com/esm-dev/esm.sh/issues/362), [#369](https://github.com/esm-dev/esm.sh/issues/369))
- Fix esm resolving before build (close [#377](https://github.com/esm-dev/esm.sh/issues/377))

## v87

- Support `?external` query, this will give you better user experience when you are using **import maps**.
  ```jsonc
  // import_map.json
  {
    "imports": {
      "preact": "https://esm.sh/preact@10.7.2",
      "preact-render-to-string": "https://esm.sh/preact-render-to-string@5.2.0?external=preact",
    }
  }
  ```
- Support `?no-dts` (equals to `?no-check`) query
- Add the 'ignore-annotations' option for esbuild ([#349](https://github.com/esm-dev/esm.sh/issues/349))
- Prevent submodules bundling their local dependencies ([#354](https://github.com/esm-dev/esm.sh/issues/354))
- Don't panic in Opera

## v86

- Support `?keep-names` query for esbuild (close [#345](https://github.com/esm-dev/esm.sh/issues/345))

## v85

- Fix `fixAliasDeps` function that imports multiple React when using `?deps=react@18,react-dom@18`

## v84

- Fix types version resolving with `?deps` query(close [#338](https://github.com/esm-dev/esm.sh/issues/338))
- Fix URL redirect with outdated build version prefix

## v83

- Replace `node-fetch` dep to `node-fetch-native` (close [#336](https://github.com/esm-dev/esm.sh/issues/336))
- Add `--keep-names` option for esbuild by default (close [#335](https://github.com/esm-dev/esm.sh/issues/335))
- Fix incorrect types with `?alias` query

## v82

- fix types with `?deps` query (close [#333](https://github.com/esm-dev/esm.sh/issues/333))

## v81

- fix `?deps` and `?alias` depth query

## v80

- Fix build error in v79

## v79

- Use `esm.sh` instead of `cdn.esm.sh`
- User semver versioning for the `x-typescript-types` header
- Fix aliasing dependencies doesn't affect typescript declaration (close [#102](https://github.com/esm-dev/esm.sh/issues/102))
- Fix using arguments in arrow function [#322](https://github.com/esm-dev/esm.sh/pull/322)
- Fix Deno check precluding esm.sh to start [#327](https://github.com/esm-dev/esm.sh/pull/327)

## v78

- Reduce database store structure
- Fix missed `renderToReadableStream` export of `react/server` in deno
- Fix `fetchPackageInfo` dead loop (close [#301](https://github.com/esm-dev/esm.sh/issues/301))
- Upgrade esbuild to **0.14.36**

## v77

- Use the latest version of `deno/std/node` from `cdn.deno.land` automatically
- Add `es2022` target
- Upgrade esbuild to **0.14.34**

## v76

- Fix `?deps` mode

## v75

- Fix types build version ignore `?pin` (close [#292](https://github.com/esm-dev/esm.sh/issues/292))
- Infect `?deps` and `?alias` to dependencies (close [#235](https://github.com/esm-dev/esm.sh/issues/235))
- Bundle `?worker` by default
- Upgrade semver to **v3** (close [#297](https://github.com/esm-dev/esm.sh/issues/297))
- Upgrade esbuild to **0.14.31**
- Upgrade deno std to **0.133.0** (close [#298](https://github.com/esm-dev/esm.sh/issues/298))

## v74

- Support `?no-require` flag, with this option you can ignore the `require(...)` call in ESM packages. To support logic like below:
  ```ts
  // index.mjs

  let depMod;
  try {
    depMod = await import("/path")
  } finally {
    // `?no-require` will skip next line when resolving
    depMod = require("/path")
  }
  ```

## v73

- Fix types dependency path (close [#287](https://github.com/esm-dev/esm.sh/issues/287))

## v72

- Support `jsx-runtime` with query: `https://esm.sh/react?pin=v72/jsx-runtime` -> `https://esm.sh/react/jsx-runtime?pin=v72`
- Support pure types package (close [#284](https://github.com/esm-dev/esm.sh/issues/284))

## v71

- Fix version resolving of dts transformer (close [#274](https://github.com/esm-dev/esm.sh/issues/274))

## v70

- Return `bare` code when `target` and `pin` provided to reduce requests
  ```js
  // https://esm.sh/react@17.0.2
  export * from "https://cdn.esm.sh/v69/react@17.0.2/es2021/react.js";
  ```
  ```js
  // https://esm.sh/react@17.0.2?target=es2020&pin=v70
  {content just from https://cdn.esm.sh/v69/react@17.0.2/es2021/react.js}
  ```
- Rollback `parseCJSModuleExports` function to v68 (close [#277](https://github.com/esm-dev/esm.sh/issues/277), [#279](https://github.com/esm-dev/esm.sh/issues/279))
- Fix `exports` resolving in package.json (close [#278](https://github.com/esm-dev/esm.sh/issues/278), [#280](https://github.com/esm-dev/esm.sh/issues/280))
- Upgrade deno `std/node` to **0.130.0**

## v69

- Force the dependency version of react equals to react-dom's version
  ```
  before: react-dom@18-rc.2 -> react@18-rc.2-next.xxxx
  now: react-dom@18-rc.2 -> react@18-rc.2
  ```
- Fix version check for prerelease (can't resolve `react` in `react-dom@rc`)
- Improve cjs module transform (can handle more edge cases, for example react-18-rc defines non-esm for browsers and deno)

## v68

- Fix `bundle` mode (close [#271](https://github.com/esm-dev/esm.sh/issues/271))
- Support `jsnext:main` in package.json (close [#272](https://github.com/esm-dev/esm.sh/issues/272))
- Improve `cjs-esm-exports` to support `UMD` format
  ```
  // exports: ['foo']
  const { exports } = parse('index.cjs', `
    (function (global, factory) {
      typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
      typeof define === 'function' && define.amd ? define(['exports'], factory) :
      (factory((global.MMDParser = global.MMDParser || {})));
    }(this, function (exports) {
      exports.foo = "bar";
    }))
  `);
  ```
- Upgrade deno node polyfill to **0.128.0**

## v67

- Force `react/jsx-dev-runtime` and `react-refresh` into **dev** mode
- Replace `typeof window !== undefined` to `typeof document !== undefined` for `deno` target
- Replace `object-assign` with `Object.assign`
- Upgrade `esbuild` to **0.14.25**

## v66

- Improve `exports` resloving of `package.json` (close [#179](https://github.com/esm-dev/esm.sh/issues/179))

## v65

- **Feature**: Support `?path` query to specify the `submodule`, this is friendly for **import maps** with options (close [#260](https://github.com/esm-dev/esm.sh/issues/260))
  ```jsonc
  // import-map.json
  {
    imports: {
      "react-dom/": "https://esm.sh/react-dom?target=es2015&path=/"
    }
  }
  ```
	```ts
	// equals to https://esm.sh/react-dom/server?target=es2015
	import { renderToString } from "react-dom/server"
	```
- Upgrade `deno.land/std/node` polyfill to **0.125.0**
- Upgrade `esbuild` to **v0.14.18**
- bugfixs for [#251](https://github.com/esm-dev/esm.sh/issues/251), [#256](https://github.com/esm-dev/esm.sh/issues/256), [#261](https://github.com/esm-dev/esm.sh/issues/261),[#262](https://github.com/esm-dev/esm.sh/issues/262)

## v64

- Fix Node.js `process` compatibility (close [#253](https://github.com/esm-dev/esm.sh/issues/253))
- Upgrade `deno.land/std/node` polyfill to **0.122.0**

## v63

- Add fs polyfill(fake) for browsers (close [#250](https://github.com/esm-dev/esm.sh/issues/250))
- Upgrade `deno.land/std/node` polyfill to **0.121.0**

## v62

- bugfixs for [#240](https://github.com/esm-dev/esm.sh/issues/240), [#242](https://github.com/esm-dev/esm.sh/issues/242), [#248](https://github.com/esm-dev/esm.sh/issues/248)


## v61

- Support **React new JSX transforms** in Deno(1.16+) (close [#225](https://github.com/esm-dev/esm.sh/issues/225))

## v60

- bugfixs for [#244](https://github.com/esm-dev/esm.sh/issues/244), [#245](https://github.com/esm-dev/esm.sh/issues/245), [#246](https://github.com/esm-dev/esm.sh/issues/246)

## v59

- improve `cjs-esm-exports` to support some edge cases:
  ```js
	var foo = exports.foo || (exports.foo = {});
	((bar) => { ... })(exports.bar || (exports.bar = {}));
	```
- Fix `@types` verisoning:
  - `marked` -> `@types/marked@4.0.1`
  - `marked@2` -> `@types/marked@2.0.5`
  - `marked?dep=@types/marked@4.0.0` -> `@types/marked@4.0.0`
- Upgrade `deno.land/std/node` polyfill to **0.119.0**
- Upgrade `esbuild` to **v0.14.8**

## v58

- Recover the stable queue
- Filter invalid pathnames like `/wp-admin/login.php`
- Fix `?pin` mode when build failed (close [#206](https://github.com/esm-dev/esm.sh/issues/206))

## v57

- Add `?pin` mode
- Improve build stability
- Fix `marked@4` import
- Fix invalid types hangs forever (close [#201](https://github.com/esm-dev/esm.sh/issues/201))

## v56
- `cjs-esm-exports` supports tslib `__exportStar` (close [#197](https://github.com/esm-dev/esm.sh/issues/197))
- Improve node `perf_hooks` polyfill
- Fix redeclared process polyfill (close [#195](https://github.com/esm-dev/esm.sh/issues/195))
- Fix `?worker` mode on deno ([#198](https://github.com/esm-dev/esm.sh/issues/198))
- Add `he` to `cjs-esm-exports` require mode allow list (close [#200](https://github.com/esm-dev/esm.sh/issues/200))
- Fix package css redirect link
- Upgrade **esbuild** to v0.13.12

## v55

- Add playground to write esm app online, try it on https://esm.sh?playground
- Add a better **cjs exports parser**: [cjs-esm-exports](https://www.npmjs.com/package/cjs-esm-exports)
- Support web worker
  ```js
  import editorWorker from '/monaco-editor/esm/vs/editor/editor.worker?worker'

	const worker = new editorWorker()
	```
-	Add `queue` interface
- Support **dataurl**, **.wasm** import
- Import deno polyfills from https://deno.land/std@0.113.0/node
- Fix package CSS

## v54

- Update deno polyfills from 0.106.0 to 0.110.0 ([#190](https://github.com/esm-dev/esm.sh/issues/190))
- Add deno `module` polyfill ([#164](https://github.com/esm-dev/esm.sh/issues/164))
- Fix (storage/fs_local) file path portability bug ([#158](https://github.com/esm-dev/esm.sh/issues/158))

## v53

- Add `Cache-Tag` header for CDN purge
- Add **s3** storage support ([#153](https://github.com/esm-dev/esm.sh/issues/153))
- Fix `require` replacement ([#154](https://github.com/esm-dev/esm.sh/issues/154))

## v52

- Fix types build ([#149](https://github.com/esm-dev/esm.sh/issues/149))
- Use `stream` and `events` from deno std/node ([#136](https://github.com/esm-dev/esm.sh/issues/148)) @talentlessguy
- Fix `localLRU` and allow for `memoryLRU` ([#148](https://github.com/esm-dev/esm.sh/issues/148)) @jimisaacs

## v51

- Fix build breaking change in v50 ([#131](https://github.com/esm-dev/esm.sh/issues/131)).
- Add `localLRU` **FS** layer ([#126](https://github.com/esm-dev/esm.sh/issues/126))
- Add a `Cache Interface` that is using to store temporary data like npm packages info.
- Do not try to build `/favicon.ico` ([#132](https://github.com/esm-dev/esm.sh/issues/132))
- Add lovely `pixi.js`, `three.js` and `@material-ui/core` testing by @jimisaacs ([#134](https://github.com/esm-dev/esm.sh/issues/134), [#139](https://github.com/esm-dev/esm.sh/issuGes/139)).

## v50

- Improve build performance to burn the server CPU cores! Before this, to build a module to ESM which has heavy deps maybe very slow since the single build task only uses one CPU core.
- Rewrite the **dts transformer** to get better deno types compatibility and faster transpile speed.
- Add Deno **testing CI** on Github.

## v49

- Improve the build process to fix an edge case reported in [#118](https://github.com/esm-dev/esm.sh/issues/118)
	```js
	const Parser = require('htmlparser').Parser;
	```
	esm (v48) output:
	```js
	import htmlparser2 from '/v48/htmlparser2@5.0.0/es2021/htmlparser2.js'
	const Parser = htmlparser2.Parser; // parser is undefined
	```
	the expected output was fixed in v49:
	```js
	import { Parser as htmlparser2Parser } from '/v48/htmlparser2@5.0.0/es2021/htmlparser2.js'
	const Parser = htmlparser2Parser; // parser is a class
	```
- Add more polyfills for Deno, huge thanks to @talentlessguy ([#117](https://github.com/esm-dev/esm.sh/issues/117))
  - path
  - querystring
  - url
  - timers
-	Better self-hosting options improved by @jimisaacs, super! ([#116](https://github.com/esm-dev/esm.sh/issues/116), [#119](https://github.com/esm-dev/esm.sh/issues/116), [#120](https://github.com/esm-dev/esm.sh/issues/120), [#122](https://github.com/esm-dev/esm.sh/issues/122))
- Add **Unlimted(max 1PB) Storage** to store builds and cache via NFS on esm.sh back server behind Cloudflare

## v48

- Improve **cjs-lexer** service to handle the edge case is shown below:
	```js
	function debounce() {};
	debounce.debounce = debounce;
	module.exports = debounce;
	```
	esm output:
	```js
	export { debounce } // this was missed
	export default debounce
	```
- Ignore `?target` in Deno (fix [#109](https://github.com/esm-dev/esm.sh/issues/109))
- Add **Storage Interface** to store data to anywhere (currently only support [postdb](https://github.com/postui/postdb) + local FS)

## v47

- Improve dts transformer to use cdn domain (fix [#104](https://github.com/esm-dev/esm.sh/issues/104))
- Update polyfills (fix [#105](https://github.com/esm-dev/esm.sh/issues/105))

## v46

- Split modules based on exports defines (ref [#78](https://github.com/esm-dev/esm.sh/issues/78))
- Add `cache-folder` config for `yarn add`
- Improve `resolveVersion` to support format 4.x (fix [#93](https://github.com/esm-dev/esm.sh/issues/93))
- Import initESM to support bare exports in package.json (fix [#97](https://github.com/esm-dev/esm.sh/issues/97))
- Bundle mode should respect the extra external (fix [#98](https://github.com/esm-dev/esm.sh/issues/98))
- Support node:path importing (fix [#100](https://github.com/esm-dev/esm.sh/issues/100))
- Pass `?alias` and `?deps` to deps (fix [#101](https://github.com/esm-dev/esm.sh/issues/101))
- Improve `cjs-lexer` sever (fix [#103](https://github.com/esm-dev/esm.sh/issues/103))
- Upgrade **rex** to **1.4.1**
- Upgrade **esbuild** to **0.12.24**

## V45

- Improve build performance
- Filter `cjs-moudle-lexer` server invalid exports output
- Improve `resolveVersion` function to support format like **4.x** (fix [#93](https://github.com/esm-dev/esm.sh/issues/93))
- Improve **dts** transform (fix [#95](https://github.com/esm-dev/esm.sh/issues/95))

## V44

- Add `Alias` feature ([#89](https://github.com/esm-dev/esm.sh/issues/89))
  ```javascript
  import useSWR from 'https://esm.sh/swr?alias=react:preact/compat'
  ```
  in combination with `?deps`:
  ```javascript
  import useSWR from 'https://esm.sh/swr?alias=react:preact/compat&deps=preact@10.5.14'
  ```
  The origin idea was came from [@lucacasonato](https://github.com/lucacasonato).
- Add `node` build target ([#84](https://github.com/esm-dev/esm.sh/issues/84))
- Check `exports` field to get entry point in `package.json`
- Run cjs-lexer as a server
- Upgrade **esbuild** to **0.11.18** with `es2021` build target
- Bugfixs for
[#90](https://github.com/esm-dev/esm.sh/issues/90),
[#85](https://github.com/esm-dev/esm.sh/issues/85),
[#83](https://github.com/esm-dev/esm.sh/issues/83),
[#77](https://github.com/esm-dev/esm.sh/issues/77),
[#65](https://github.com/esm-dev/esm.sh/issues/65),
[#48](https://github.com/esm-dev/esm.sh/issues/48),
[#41](https://github.com/esm-dev/esm.sh/issues/41).

## V43

- Add `/status.json` api
- Use previous build instead of waiting/404 (fix [#74](https://github.com/esm-dev/esm.sh/issues/74))
- Fix deps query ([#71](https://github.com/esm-dev/esm.sh/issues/71))

## V42

- Add `__esModule` reserved word
- Align require change for esbuild 0.12
- Fix setImmediate polyfill args ([#75](https://github.com/esm-dev/esm.sh/issues/75))
- Upgrade **esbuild** to **0.11.12**

## V41

- Add `timeout` (30 seconds) for new build request, or use previous build version instead if it exists
- Fix `bundle` mode
- Fix build dead loop
- Upgrade **esbuild** to **0.11.12**

## V40

- Update polyfills for node builtin modules
- Upgrade **esbuild** to **0.11.9**

## V39

- Imporve `parseCJSModuleExports` to support json module
- Pass `NODE_ENV` to `parseCJSModuleExports`
- Update node buffer polyfill
- Upgrade postdb to **v0.6.2**

## V38

- Fix build for packages with `module` type ([#48](https://github.com/esm-dev/esm.sh/issues/48))
- Improve `parseCJSModuleExports` function (use cjs-module-lexer and nodejs eval both to parse cjs exports, and ignore JSON module)
- Pass `NODE_ENV` to `parseCJSModuleExports` function
- Upgrade **esbuild** to **0.11.6**

## V37

- Add **bundle** mode
- Fix module exports parsing

## V36

- Fix esm build for some edge cases
- Add simple test (thanks @zhoukekestar)
- Upgrade esbuild to 0.11.5

## V35

- Set build `target` by the `user-agent` of browser automaticlly

## V34

- Remove bundle mode **&middot; Breaking**
- Add build queue instead of mutex lock
- Use AST([cjs-module-lexer](https://github.com/guybedford/cjs-module-lexer)) to parse cjs exports
- Add a testing page at https://esm.sh?test
- Fix `__setImmediate$` is not defined
- Support exports define in package.json
- Support mjs extension
- Improve NpmPackage resolve (close [#41](https://github.com/esm-dev/esm.sh/issues/41))
- Upgrade esbuild to **0.11.4**
- Upgrade rex to **1.3.0**
