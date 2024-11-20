package server

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/ije/gox/utils"
)

// BuildEntry represents the build entrypoints of a module
type BuildEntry struct {
	esm string
	cjs string
	dts string
}

// isEmpty checks if the entry is empty
func (entry *BuildEntry) isEmpty() bool {
	return entry.esm == "" && entry.cjs == "" && entry.dts == ""
}

// hasEntry checks if the entrypoint of the given type exists
func (entry *BuildEntry) hasEntry(entryType string) bool {
	switch entryType {
	case "esm":
		return entry.esm != ""
	case "cjs":
		return entry.cjs != ""
	case "dts":
		return entry.dts != ""
	}
	return false
}

// updateEntry updates the entrypoint of the given type
func (entry *BuildEntry) updateEntry(entryType string, entryPath string) {
	switch entryType {
	case "esm":
		entry.esm = entryPath
	case "cjs":
		entry.cjs = entryPath
	case "dts":
		entry.dts = entryPath
	}
}

// resolve resolves the entrypoint of the given type
func (entry *BuildEntry) resolve(ctx *BuildContext, mType string, condition interface{}, entryType string) {
	if entry.hasEntry(entryType) {
		return
	}
	if s, ok := condition.(string); ok {
		entry.updateEntry(entryType, s)
	} else if om, ok := condition.(*OrderedMap); ok {
		if v, ok := om.m["default"]; ok {
			if s, ok := v.(string); ok && s != "" {
				entry.updateEntry(entryType, s)
			}
		}
		e := ctx.resolveConditionExportEntry(om, mType)
		if e.esm != "" && !entry.hasEntry("esm") {
			entry.updateEntry("esm", e.esm)
		}
		if e.cjs != "" && !entry.hasEntry("cjs") {
			entry.updateEntry("cjs", e.cjs)
		}
		if e.dts != "" && !entry.hasEntry("dts") {
			entry.updateEntry("dts", e.dts)
		}
	}
}

func (ctx *BuildContext) Path() string {
	if ctx.path != "" {
		return ctx.path
	}

	esmPath := ctx.esmPath
	if ctx.target == "types" {
		if strings.HasSuffix(esmPath.SubPath, ".d.ts") {
			ctx.path = fmt.Sprintf(
				"/%s/%s%s",
				esmPath.PackageName(),
				ctx.getBuildArgsPrefix(true),
				esmPath.SubPath,
			)
		} else {
			ctx.path = "/" + esmPath.String()
		}
		return ctx.path
	}

	name := strings.TrimSuffix(path.Base(esmPath.PkgName), ".js")
	if esmPath.SubBareName != "" {
		if esmPath.SubBareName == name {
			// if the sub-module name is same as the package name
			name = "__" + esmPath.SubBareName
		} else {
			name = esmPath.SubBareName
		}
		// workaround for es5-ext "../#/.." path
		if esmPath.PkgName == "es5-ext" {
			name = strings.ReplaceAll(name, "/#/", "/%23/")
		}
	}

	if ctx.dev {
		name += ".development"
	}
	if ctx.bundleMode == BundleAll {
		name += ".bundle"
	} else if ctx.bundleMode == BundleFalse {
		name += ".nobundle"
	}
	ctx.path = fmt.Sprintf(
		"/%s/%s%s/%s.mjs",
		esmPath.PackageName(),
		ctx.getBuildArgsPrefix(ctx.target == "types"),
		ctx.target,
		name,
	)
	return ctx.path
}

func (ctx *BuildContext) getImportPath(esmPath ESMPath, buildArgsPrefix string) string {
	name := strings.TrimSuffix(path.Base(esmPath.PkgName), ".js")
	if esmPath.SubBareName != "" {
		if esmPath.SubBareName == name {
			// if the sub-module name is same as the package name
			name = "__" + esmPath.SubBareName
		} else {
			name = esmPath.SubBareName
		}
		// workaround for es5-ext "../#/.." path
		if esmPath.PkgName == "es5-ext" {
			name = strings.ReplaceAll(name, "/#/", "/%23/")
		}
	}
	if ctx.dev {
		name += ".development"
	}
	return fmt.Sprintf(
		"/%s/%s%s/%s.mjs",
		esmPath.PackageName(),
		buildArgsPrefix,
		ctx.target,
		name,
	)
}

func (ctx *BuildContext) getSavepath() string {
	return normalizeSavePath(ctx.zoneId, path.Join("builds", ctx.Path()))
}

func (ctx *BuildContext) getBuildArgsPrefix(isDts bool) string {
	if a := encodeBuildArgs(ctx.args, isDts); a != "" {
		return "X-" + a + "/"
	}
	return ""
}

func (ctx *BuildContext) getNodeEnv() string {
	if ctx.dev {
		return "development"
	}
	return "production"
}

func (ctx *BuildContext) isDenoTarget() bool {
	return ctx.target == "deno" || ctx.target == "denonext"
}

func (ctx *BuildContext) isBrowserTarget() bool {
	return strings.HasPrefix(ctx.target, "es")
}

func (ctx *BuildContext) existsPkgFile(fp ...string) bool {
	args := make([]string, 1+len(fp))
	args[0] = ctx.pkgDir
	copy(args[1:], fp)
	return existsFile(path.Join(args...))
}

func (ctx *BuildContext) lookupDep(specifier string, isDts bool) (esm ESMPath, packageJson *PackageJSON, err error) {
	pkgName, version, subpath, _ := splitESMPath(specifier)
lookup:
	if v, ok := ctx.args.deps[pkgName]; ok {
		packageJson, err = ctx.npmrc.getPackageInfo(pkgName, v)
		if err == nil {
			esm = ESMPath{
				PkgName:     pkgName,
				PkgVersion:  packageJson.Version,
				SubPath:     subpath,
				SubBareName: toModuleBareName(subpath, true),
			}
		}
		return
	}
	pkgJsonPath := path.Join(ctx.wd, "node_modules", ".pnpm", "node_modules", pkgName, "package.json")
	if !existsFile(pkgJsonPath) {
		pkgJsonPath = path.Join(ctx.wd, "node_modules", pkgName, "package.json")
	}
	var rawInfo PackageJSONRaw
	if existsFile(pkgJsonPath) && utils.ParseJSONFile(pkgJsonPath, &rawInfo) == nil {
		esm = ESMPath{
			PkgName:     pkgName,
			PkgVersion:  rawInfo.Version,
			SubPath:     subpath,
			SubBareName: toModuleBareName(subpath, true),
		}
		packageJson = rawInfo.ToNpmPackage()
		return
	}
	if version == "" {
		if v, ok := ctx.packageJson.Dependencies[pkgName]; ok {
			if strings.HasPrefix(v, "npm:") {
				pkgName, version, _, _ = splitESMPath(v[4:])
			} else {
				version = v
			}
		} else if v, ok = ctx.packageJson.PeerDependencies[pkgName]; ok {
			if strings.HasPrefix(v, "npm:") {
				pkgName, version, _, _ = splitESMPath(v[4:])
			} else {
				version = v
			}
		} else {
			version = "latest"
		}
	}
	packageJson, err = ctx.npmrc.getPackageInfo(pkgName, version)
	if err == nil {
		esm = ESMPath{
			PkgName:     pkgName,
			PkgVersion:  packageJson.Version,
			SubPath:     subpath,
			SubBareName: toModuleBareName(subpath, true),
		}
	}
	if err != nil && strings.HasSuffix(err.Error(), " not found") && isDts && !strings.HasPrefix(pkgName, "@types/") {
		pkgName = toTypesPackageName(pkgName)
		goto lookup
	}
	return
}

func (ctx *BuildContext) resolveEntry(esm ESMPath) (entry BuildEntry) {
	pkgDir := ctx.pkgDir

	if esm.SubBareName != "" {
		if endsWith(esm.SubPath, ".d.ts", ".d.mts", ".d.cts") {
			entry.dts = normalizeEntryPath(esm.SubPath)
			return
		}

		if endsWith(esm.SubPath, ".jsx", ".ts", ".tsx", ".mts", ".svelte", ".vue") {
			entry.esm = normalizeEntryPath(esm.SubPath)
			return
		}

		subModuleName := esm.SubBareName

		// reslove sub-module using `exports` conditions if exists
		// see https://nodejs.org/api/packages.html#package-entry-points
		if ctx.packageJson.Exports != nil {
			exportEntry := BuildEntry{}
			if om, ok := ctx.packageJson.Exports.(*OrderedMap); ok {
				for _, name := range om.keys {
					conditions := om.Get(name)
					if name == "./"+subModuleName || stripModuleExt(name, ".js", ".cjs", ".mjs") == "./"+subModuleName {
						if s, ok := conditions.(string); ok {
							/**
							exports: {
								"./lib/foo": "./lib/foo.js"
							}
							*/
							if ctx.packageJson.Type == "module" {
								exportEntry.esm = s
							} else {
								exportEntry.cjs = s
							}
						} else if om, ok := conditions.(*OrderedMap); ok {
							/**
							exports: {
								"./lib/foo": {
									"require": "./lib/foo.js",
									"import": "./esm/foo.js",
									"types": "./types/foo.d.ts"
								}
							}
							*/
							exportEntry = ctx.resolveConditionExportEntry(om, ctx.packageJson.Type)
						}
						break
					} else if diff, ok := matchAsteriskExports(name, esm); ok {
						if s, ok := conditions.(string); ok {
							/**
							exports: {
								"./lib/foo/*": "./lib/foo/*.js",
							}
							*/
							e := strings.ReplaceAll(s, "*", diff)
							if ctx.packageJson.Type == "module" {
								exportEntry.esm = e
							} else {
								exportEntry.cjs = e
							}
						} else if om, ok := conditions.(*OrderedMap); ok {
							/**
							exports: {
								"./lib/foo/*": {
									"require": "./lib/foo/*.js",
									"import": "./esm/lib/foo/*.js",
									"types": "./types/foo/*.d.ts"
								},
							}
							*/
							exportEntry = ctx.resolveConditionExportEntry(resloveAsteriskPathMapping(om, diff), ctx.packageJson.Type)
						}
					}
				}
			}
			normalizeBuildEntry(ctx, &exportEntry)
			if exportEntry.esm != "" && ctx.existsPkgFile(exportEntry.esm) {
				entry.esm = exportEntry.esm
			}
			if exportEntry.cjs != "" && ctx.existsPkgFile(exportEntry.cjs) {
				entry.cjs = exportEntry.cjs
			}
			if exportEntry.dts != "" && ctx.existsPkgFile(exportEntry.dts) {
				entry.dts = exportEntry.dts
			}
		}

		var rawInfo PackageJSONRaw
		if utils.ParseJSONFile(path.Join(pkgDir, subModuleName, "package.json"), &rawInfo) == nil {
			p := rawInfo.ToNpmPackage()
			if entry.esm == "" && p.Module != "" {
				entry.esm = "./" + path.Join(subModuleName, p.Module)
			}
			if entry.cjs == "" && p.Main != "" {
				entry.cjs = "./" + path.Join(subModuleName, p.Main)
			}
			if entry.dts == "" {
				if p.Types != "" {
					entry.dts = "./" + path.Join(subModuleName, p.Types)
				} else if p.Typings != "" {
					entry.dts = "./" + path.Join(subModuleName, p.Typings)
				}
			}
		}

		if entry.esm == "" {
			if ctx.existsPkgFile(subModuleName + ".mjs") {
				entry.esm = "./" + subModuleName + ".mjs"
			} else if ctx.existsPkgFile(subModuleName, "index.mjs") {
				entry.esm = "./" + subModuleName + "/index.mjs"
			} else if ctx.packageJson.Type == "module" {
				if ctx.existsPkgFile(subModuleName + ".js") {
					entry.esm = "./" + subModuleName + ".js"
				} else if ctx.existsPkgFile(subModuleName, "index.js") {
					entry.esm = "./" + subModuleName + "/index.js"
				}
			}
		}

		if entry.cjs == "" {
			if ctx.existsPkgFile(subModuleName + ".cjs") {
				entry.cjs = "./" + subModuleName + ".cjs"
			} else if ctx.existsPkgFile(subModuleName, "index.cjs") {
				entry.cjs = "./" + subModuleName + "/index.cjs"
			} else if ctx.packageJson.Type != "module" {
				if ctx.existsPkgFile(subModuleName + ".js") {
					entry.cjs = "./" + subModuleName + ".js"
				} else if ctx.existsPkgFile(subModuleName, "index.js") {
					entry.cjs = "./" + subModuleName + "/index.js"
				}
			}
		}

		if entry.dts == "" {
			if entry.esm != "" && ctx.existsPkgFile(stripModuleExt(entry.esm)+".d.ts") {
				entry.dts = stripModuleExt(entry.esm) + ".d.ts"
			} else if entry.cjs != "" && ctx.existsPkgFile(stripModuleExt(entry.cjs)+".d.ts") {
				entry.dts = stripModuleExt(entry.cjs) + ".d.ts"
			} else if ctx.existsPkgFile(subModuleName + ".d.mts") {
				entry.dts = "./" + subModuleName + ".d.mts"
			} else if ctx.existsPkgFile(subModuleName + ".d.ts") {
				entry.dts = "./" + subModuleName + ".d.ts"
			} else if ctx.existsPkgFile(subModuleName, "index.d.mts") {
				entry.dts = "./" + subModuleName + "/index.d.mts"
			} else if ctx.existsPkgFile(subModuleName, "index.d.ts") {
				entry.dts = "./" + subModuleName + "/index.d.ts"
			}
		}
	} else {
		entry = BuildEntry{
			esm: ctx.packageJson.Module,
			cjs: ctx.packageJson.Main,
			dts: ctx.packageJson.Types,
		}
		if entry.dts == "" && ctx.packageJson.Typings != "" {
			entry.dts = ctx.packageJson.Typings
		}

		if exports := ctx.packageJson.Exports; exports != nil {
			exportEntry := BuildEntry{}
			if om, ok := exports.(*OrderedMap); ok {
				v, ok := om.m["."]
				if ok {
					if s, ok := v.(string); ok {
						/**
						exports: {
							".": "./index.js"
						}
						*/
						if ctx.packageJson.Type == "module" {
							exportEntry.esm = s
						} else {
							exportEntry.cjs = s
						}
					} else if om, ok := v.(*OrderedMap); ok {
						/**
						exports: {
							".": {
								"require": "./cjs/index.js",
								"import": "./esm/index.js"
							}
						}
						*/
						exportEntry = ctx.resolveConditionExportEntry(om, ctx.packageJson.Type)
					}
				} else {
					/**
					exports: {
						"require": "./cjs/index.js",
						"import": "./esm/index.js"
					}
					*/
					exportEntry = ctx.resolveConditionExportEntry(om, ctx.packageJson.Type)
				}
			} else if s, ok := exports.(string); ok {
				/**
				exports: "./index.js"
				*/
				if ctx.packageJson.Type == "module" {
					exportEntry.esm = s
				} else {
					exportEntry.cjs = s
				}
			}
			normalizeBuildEntry(ctx, &exportEntry)
			if exportEntry.esm != "" && ctx.existsPkgFile(exportEntry.esm) {
				entry.esm = exportEntry.esm
			}
			if exportEntry.cjs != "" && ctx.existsPkgFile(exportEntry.cjs) {
				entry.cjs = exportEntry.cjs
			}
			if exportEntry.dts != "" && ctx.existsPkgFile(exportEntry.dts) {
				entry.dts = exportEntry.dts
			}
		}

		if entry.esm == "" {
			if ctx.packageJson.Type == "module" && ctx.existsPkgFile("index.js") {
				entry.esm = "./index.js"
			} else if ctx.existsPkgFile("index.mjs") {
				entry.esm = "./index.mjs"
			}
		}

		if entry.cjs == "" {
			if ctx.packageJson.Type != "module" && ctx.existsPkgFile("index.js") {
				entry.cjs = "./index.js"
			} else if ctx.existsPkgFile("index.cjs") {
				entry.cjs = "./index.cjs"
			}
		}

		if entry.dts == "" {
			if ctx.existsPkgFile("index.d.mts") {
				entry.dts = "./index.d.mts"
			} else if ctx.existsPkgFile("index.d.ts") {
				entry.dts = "./index.d.ts"
			}
		}
		if entry.dts == "" && entry.esm != "" {
			if ctx.existsPkgFile(stripModuleExt(entry.esm) + ".d.mts") {
				entry.dts = stripModuleExt(entry.esm) + ".d.mts"
			} else if ctx.existsPkgFile(stripModuleExt(entry.esm) + ".d.ts") {
				entry.dts = stripModuleExt(entry.esm) + ".d.ts"
			} else if stripModuleExt(path.Base(entry.esm)) == "index" {
				dir, _ := utils.SplitByLastByte(entry.esm, '/')
				if ctx.existsPkgFile(dir, "index.d.mts") {
					entry.dts = dir + "/index.d.mts"
				} else if ctx.existsPkgFile(dir, "index.d.ts") {
					entry.dts = dir + "/index.d.ts"
				}
			}
		}
		if entry.dts == "" && entry.cjs != "" {
			if ctx.existsPkgFile(stripModuleExt(entry.cjs) + ".d.mts") {
				entry.dts = stripModuleExt(entry.cjs) + ".d.mts"
			} else if ctx.existsPkgFile(stripModuleExt(entry.cjs) + ".d.ts") {
				entry.dts = stripModuleExt(entry.cjs) + ".d.ts"
			} else if stripModuleExt(path.Base(entry.cjs)) == "index" {
				dir, _ := utils.SplitByLastByte(entry.cjs, '/')
				if ctx.existsPkgFile(dir, "index.d.mts") {
					entry.dts = dir + "/index.d.mts"
				} else if ctx.existsPkgFile(dir, "index.d.ts") {
					entry.dts = dir + "/index.d.ts"
				}
			}
		}
	}

	// resovle dts from `typesVersions` field if it's defined
	// see https://www.typescriptlang.org/docs/handbook/declaration-files/publishing.html#version-selection-with-typesversions
	if typesVersions := ctx.packageJson.TypesVersions; len(typesVersions) > 0 && entry.dts != "" {
		versions := make(sort.StringSlice, len(typesVersions))
		i := 0
		for c := range typesVersions {
			if strings.HasPrefix(c, ">") {
				versions[i] = c
				i++
			}
		}
		versions = versions[:i]
		if versions.Len() > 0 {
			versions.Sort()
			latestVersion := typesVersions[versions[versions.Len()-1]]
			if mapping, ok := latestVersion.(map[string]interface{}); ok {
				var paths interface{}
				var matched bool
				var exact bool
				var suffix string
				dts := normalizeEntryPath(entry.dts)
				paths, matched = mapping[dts]
				if !matched {
					// try to match the dts wihout leading "./"
					paths, matched = mapping[strings.TrimPrefix(dts, "./")]
				}
				if matched {
					exact = true
				}
				if !matched {
					for key, value := range mapping {
						if strings.HasSuffix(key, "/*") {
							key = normalizeEntryPath(key)
							if strings.HasPrefix(dts, strings.TrimSuffix(key, "/*")) {
								paths = value
								matched = true
								suffix = strings.TrimPrefix(dts, strings.TrimSuffix(key, "*"))
								break
							}
						}
					}
				}
				if !matched {
					paths, matched = mapping["*"]
				}
				if matched {
					if a, ok := paths.([]interface{}); ok && len(a) > 0 {
						if path, ok := a[0].(string); ok {
							path = normalizeEntryPath(path)
							if exact {
								entry.dts = path
							} else {
								prefix, _ := utils.SplitByLastByte(path, '*')
								if suffix != "" {
									entry.dts = prefix + suffix
								} else if strings.HasPrefix(dts, prefix) {
									diff := strings.TrimPrefix(dts, prefix)
									entry.dts = strings.ReplaceAll(path, "*", diff)
								} else {
									entry.dts = prefix + dts[2:]
								}
							}
						}
					}
				}
			}
		}
	}

	// check the `browser` field
	if len(ctx.packageJson.Browser) > 0 && ctx.isBrowserTarget() {
		// normalize the entry
		normalizeBuildEntry(ctx, &entry)
		if entry.esm != "" {
			m, ok := ctx.packageJson.Browser[entry.esm]
			if ok && isRelPathSpecifier(m) {
				entry.esm = m
			}
		}
		if entry.cjs != "" {
			m, ok := ctx.packageJson.Browser[entry.cjs]
			if ok && isRelPathSpecifier(m) {
				entry.cjs = m
			}
		}
		if esm.SubBareName == "" {
			if m, ok := ctx.packageJson.Browser["."]; ok && isRelPathSpecifier(m) {
				if strings.HasSuffix(m, ".mjs") {
					entry.esm = m
				} else if entry.esm == "" {
					entry.cjs = m
				}
			}
		}
	}

	// normalize the entry
	normalizeBuildEntry(ctx, &entry)
	return
}

// see https://nodejs.org/api/packages.html#nested-conditions
func (ctx *BuildContext) resolveConditionExportEntry(conditions *OrderedMap, mType string) (entry BuildEntry) {
	entryKey := "esm"
	switch mType {
	case "", "commonjs":
		entryKey = "cjs"
	case "module":
		entryKey = "esm"
	case "types":
		entryKey = "dts"
	}

	if len(ctx.args.conditions) > 0 {
		for _, conditionName := range ctx.args.conditions {
			condition := conditions.Get(conditionName)
			if condition != nil {
				entry.resolve(ctx, mType, condition, entryKey)
			}
		}
	}

	if ctx.dev {
		condition := conditions.Get("development")
		if condition != nil {
			entry.resolve(ctx, mType, condition, entryKey)
		}
	}

	if ctx.isBrowserTarget() {
		condition := conditions.Get("browser")
		if condition != nil {
			entry.resolve(ctx, mType, condition, entryKey)
		}
	} else if ctx.isDenoTarget() {
		var condition interface{}
		for _, conditionName := range []string{"deno", "node"} {
			condition = conditions.Get(conditionName)
			if condition != nil {
				// entry.ibc = conditionName != "browser"
				entry.resolve(ctx, mType, condition, entryKey)
				break
			}
		}
	} else if ctx.target == "node" {
		condition := conditions.Get("node")
		if condition != nil {
			entry.resolve(ctx, mType, condition, entryKey)
		}
	}

	for _, conditionName := range conditions.keys {
		condition := conditions.Get(conditionName)
		switch conditionName {
		case "module", "import", "es2015":
			entry.resolve(ctx, "module", condition, "esm")
		case "require":
			entry.resolve(ctx, "commonjs", condition, "cjs")
		case "types", "typings":
			entry.resolve(ctx, "types", condition, "dts")
		case "default":
			entry.resolve(ctx, mType, condition, entryKey)
		}
	}
	return
}

func (ctx *BuildContext) resolveExternalModule(specifier string, kind api.ResolveKind) (resolvedPath string, err error) {
	defer func() {
		if err == nil {
			resolvedPathFull := resolvedPath
			// use relative path for sub-module of current package
			if strings.HasPrefix(specifier, ctx.packageJson.Name+"/") {
				rp, err := relPath(path.Dir(ctx.Path()), resolvedPath)
				if err == nil {
					resolvedPath = rp
				}
			}
			// mark the resolved path for _preload_
			if kind != api.ResolveJSDynamicImport {
				ctx.importMap = append(ctx.importMap, [2]string{resolvedPathFull, resolvedPath})
			}
			// if it's `require("module")` call
			if kind == api.ResolveJSRequireCall {
				ctx.cjsRequires = append(ctx.cjsRequires, [3]string{specifier, resolvedPathFull, resolvedPath})
				resolvedPath = specifier
			}
		}
	}()

	// it's the entry of current package from GitHub
	if npm := ctx.packageJson; ctx.esmPath.GhPrefix && (specifier == npm.Name || specifier == npm.PkgName) {
		resolvedPath = ctx.getImportPath(ESMPath{
			PkgName:    npm.Name,
			PkgVersion: npm.Version,
			GhPrefix:   true,
		}, ctx.getBuildArgsPrefix(false))
		return
	}

	// node builtin module
	if isNodeBuiltInModule(specifier) {
		if ctx.args.externalAll || ctx.target == "node" || ctx.target == "denonext" || ctx.args.external.Has(specifier) {
			resolvedPath = specifier
		} else if ctx.target == "deno" {
			resolvedPath = fmt.Sprintf("https://deno.land/std@0.177.1/node/%s.ts", specifier[5:])
		} else {
			resolvedPath = fmt.Sprintf("/node/%s.mjs", specifier[5:])
		}
		return
	}

	// check `?external`
	if ctx.args.externalAll || ctx.args.external.Has(toPackageName(specifier)) {
		resolvedPath = specifier
		return
	}

	// it's a sub-module of current package
	if strings.HasPrefix(specifier, ctx.packageJson.Name+"/") {
		subPath := strings.TrimPrefix(specifier, ctx.packageJson.Name+"/")
		subModule := ESMPath{
			PkgName:     ctx.esmPath.PkgName,
			PkgVersion:  ctx.esmPath.PkgVersion,
			SubPath:     subPath,
			SubBareName: toModuleBareName(subPath, false),
			GhPrefix:    ctx.esmPath.GhPrefix,
		}
		if ctx.subBuilds != nil {
			b := &BuildContext{
				zoneId:      ctx.zoneId,
				npmrc:       ctx.npmrc,
				esmPath:     subModule,
				packageJson: ctx.packageJson,
				deprecated:  ctx.deprecated,
				args:        ctx.args,
				target:      ctx.target,
				dev:         ctx.dev,
				wd:          ctx.wd,
				pkgDir:      ctx.pkgDir,
				pnpmPkgDir:  ctx.pnpmPkgDir,
				subBuilds:   ctx.subBuilds,
			}
			if ctx.bundleMode == BundleFalse {
				b.bundleMode = BundleFalse
			}
			pathname := b.Path()
			if !ctx.subBuilds.Has(pathname) {
				ctx.subBuilds.Add(pathname)
				ctx.wg.Add(1)
				go func() {
					defer ctx.wg.Done()
					b.Build()
				}()
			}
		}
		resolvedPath = ctx.getImportPath(subModule, ctx.getBuildArgsPrefix(false))
		if ctx.bundleMode == BundleFalse {
			n, e := utils.SplitByLastByte(resolvedPath, '.')
			resolvedPath = n + ".nobundle." + e
		}
		return
	}

	// common npm dependency
	pkgName, version, subpath, _ := splitESMPath(specifier)
	if version == "" {
		if pkgName == ctx.esmPath.PkgName {
			version = ctx.esmPath.PkgVersion
		} else if pkgVerson, ok := ctx.args.deps[pkgName]; ok {
			version = pkgVerson
		} else if v, ok := ctx.packageJson.Dependencies[pkgName]; ok {
			version = strings.TrimSpace(v)
		} else if v, ok := ctx.packageJson.PeerDependencies[pkgName]; ok {
			version = strings.TrimSpace(v)
		} else {
			version = "latest"
		}
	}

	// force the version of 'react' (as dependency) equals to 'react-dom'
	if ctx.esmPath.PkgName == "react-dom" && pkgName == "react" {
		version = ctx.esmPath.PkgVersion
	}

	depPath := ESMPath{
		PkgName:     pkgName,
		PkgVersion:  version,
		SubPath:     subpath,
		SubBareName: toModuleBareName(subpath, true),
	}

	// resolve alias in dependencies
	// follow https://docs.npmjs.com/cli/v10/configuring-npm/package-json#git-urls-as-dependencies
	// e.g. "@mark/html": "npm:@jsr/mark__html@^1.0.0"
	// e.g. "tslib": "git+https://github.com/microsoft/tslib.git#v2.3.0"
	// e.g. "react": "github:facebook/react#v18.2.0"
	{
		// ban file specifier
		if strings.HasPrefix(version, "file:") {
			resolvedPath = fmt.Sprintf("/error.js?type=unsupported-file-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
			return
		}
		if isHttpSepcifier(version) {
			u, e := url.Parse(version)
			if e != nil || u.Host != "pkg.pr.new" {
				resolvedPath = fmt.Sprintf("/error.js?type=unsupported-http-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
				return
			}
			pkgName, rest := utils.SplitByLastByte(u.Path[1:], '@')
			if rest == "" {
				resolvedPath = fmt.Sprintf("/error.js?type=invalid-http-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
				return
			}
			version, _ := utils.SplitByFirstByte(rest, '/')
			if version == "" || !regexpVersion.MatchString(version) {
				resolvedPath = fmt.Sprintf("/error.js?type=invalid-http-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
				return
			}
			depPath.PkgName = pkgName
			depPath.PkgVersion = version
			depPath.PrPrefix = true
		} else if strings.HasPrefix(version, "npm:") {
			depPath.PkgName, depPath.PkgVersion, _, _ = splitESMPath(version[4:])
		} else if strings.HasPrefix(version, "jsr:") {
			pkgName, pkgVersion, _, _ := splitESMPath(version[4:])
			if !strings.HasPrefix(pkgName, "@") || !strings.ContainsRune(pkgName, '/') {
				resolvedPath = fmt.Sprintf("/error.js?type=invalid-jsr-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
				return
			}
			scope, name := utils.SplitByFirstByte(pkgName, '/')
			depPath.PkgName = "@jsr/" + scope[1:] + "__" + name
			depPath.PkgVersion = pkgVersion
		} else if strings.HasPrefix(version, "git+ssh://") || strings.HasPrefix(version, "git+https://") || strings.HasPrefix(version, "git://") {
			gitUrl, e := url.Parse(version)
			if e != nil || gitUrl.Hostname() != "github.com" {
				resolvedPath = fmt.Sprintf("/error.js?type=unsupported-git-dependency&name=%s&importer=%s", pkgName, ctx.esmPath)
				return
			}
			repo := strings.TrimSuffix(gitUrl.Path[1:], ".git")
			if gitUrl.Scheme == "git+ssh" {
				repo = gitUrl.Port() + "/" + repo
			}
			depPath.PkgName = repo
			depPath.PkgVersion = strings.TrimPrefix(url.QueryEscape(gitUrl.Fragment), "semver:")
			depPath.GhPrefix = true
		} else if strings.HasPrefix(version, "github:") || (!strings.HasPrefix(version, "@") && strings.ContainsRune(version, '/')) {
			repo, fragment := utils.SplitByLastByte(strings.TrimPrefix(version, "github:"), '#')
			depPath.GhPrefix = true
			depPath.PkgName = repo
			depPath.PkgVersion = strings.TrimPrefix(url.QueryEscape(fragment), "semver:")
		}
	}

	// fetch the latest tag as the version of the repository
	if depPath.GhPrefix && depPath.PkgVersion == "" {
		var refs []GitRef
		refs, err = listRepoRefs(fmt.Sprintf("https://github.com/%s", depPath.PkgName))
		if err != nil {
			return
		}
		for _, ref := range refs {
			if ref.Ref == "HEAD" {
				depPath.PkgVersion = ref.Sha[:16]
				break
			}
		}
	}

	var isFixedVersion bool
	if depPath.GhPrefix {
		isFixedVersion = isCommitish(depPath.PkgVersion) || regexpVersionStrict.MatchString(strings.TrimPrefix(depPath.PkgVersion, "v"))
	} else if depPath.PrPrefix {
		isFixedVersion = true
	} else {
		isFixedVersion = regexpVersionStrict.MatchString(depPath.PkgVersion)
	}
	args := BuildArgs{
		alias:      ctx.args.alias,
		deps:       ctx.args.deps,
		external:   ctx.args.external,
		conditions: ctx.args.conditions,
		exports:    NewStringSet(),
	}

	err = normalizeBuildArgs(ctx.npmrc, ctx.wd, &args, depPath)
	if err != nil {
		return
	}

	if isFixedVersion {
		buildArgsPrefix := ""
		if a := encodeBuildArgs(args, false); a != "" {
			buildArgsPrefix = "X-" + a + "/"
		}
		resolvedPath = ctx.getImportPath(depPath, buildArgsPrefix)
		return
	}

	if strings.ContainsRune(depPath.PkgVersion, '|') || strings.ContainsRune(depPath.PkgVersion, ' ') {
		// fetch the latest version of the package based on the semver range
		var p *PackageJSON
		_, p, err = ctx.lookupDep(pkgName+"@"+version, false)
		if err != nil {
			return
		}
		depPath.PkgVersion = "^" + p.Version
	}

	resolvedPath = "/" + depPath.String()
	// workaround for es5-ext "../#/.." path
	if depPath.PkgName == "es5-ext" {
		resolvedPath = strings.ReplaceAll(resolvedPath, "/#/", "/%23/")
	}
	params := []string{}
	if len(args.alias) > 0 {
		var alias []string
		for k, v := range args.alias {
			alias = append(alias, fmt.Sprintf("%s:%s", k, v))
		}
		params = append(params, "alias="+strings.Join(alias, ","))
	}
	if len(args.deps) > 0 {
		var deps sort.StringSlice
		for n, v := range args.deps {
			deps = append(deps, n+"@"+v)
		}
		deps.Sort()
		params = append(params, "deps="+strings.Join(deps, ","))
	}
	if args.external.Len() > 0 {
		external := make(sort.StringSlice, args.external.Len())
		for i, e := range args.external.Values() {
			external[i] = e
		}
		external.Sort()
		params = append(params, "external="+strings.Join(external, ","))
	}
	if len(args.conditions) > 0 {
		conditions := make(sort.StringSlice, len(args.conditions))
		copy(conditions, args.conditions)
		conditions.Sort()
		params = append(params, "conditions="+strings.Join(conditions, ","))
	}
	if ctx.pinedTarget {
		params = append(params, "target="+ctx.target)
	}
	if ctx.dev {
		params = append(params, "dev")
	}
	if strings.HasSuffix(resolvedPath, ".json") {
		params = append(params, "module")
	}
	if len(params) > 0 {
		resolvedPath += "?" + strings.Join(params, "&")
	}
	return
}

func (ctx *BuildContext) resloveDTS(entry BuildEntry) (string, error) {
	if entry.dts != "" {
		if !ctx.existsPkgFile(entry.dts) {
			return "", nil
		}
		return fmt.Sprintf(
			"/%s/%s%s",
			ctx.esmPath.PackageName(),
			ctx.getBuildArgsPrefix(true),
			strings.TrimPrefix(entry.dts, "./"),
		), nil
	}

	if ctx.esmPath.SubPath != "" && (ctx.packageJson.Types != "" || ctx.packageJson.Typings != "") {
		return "", nil
	}

	// lookup types in @types scope
	if packageJson := ctx.packageJson; packageJson.Types == "" && !strings.HasPrefix(packageJson.Name, "@types/") && regexpVersionStrict.MatchString(packageJson.Version) {
		versionParts := strings.Split(packageJson.Version, ".")
		versions := []string{
			versionParts[0] + "." + versionParts[1], // major.minor
			versionParts[0],                         // major
		}
		typesPkgName := toTypesPackageName(packageJson.Name)
		pkgVersion, ok := ctx.args.deps[typesPkgName]
		if ok {
			// use the version of the `?deps` query if it exists
			versions = append([]string{pkgVersion}, versions...)
		}
		for _, version := range versions {
			p, err := ctx.npmrc.getPackageInfo(typesPkgName, version)
			if err == nil {
				dtsModule := ESMPath{
					PkgName:     typesPkgName,
					PkgVersion:  p.Version,
					SubPath:     ctx.esmPath.SubPath,
					SubBareName: ctx.esmPath.SubBareName,
				}
				b := NewBuildContext(ctx.zoneId, ctx.npmrc, dtsModule, ctx.args, "types", false, BundleFalse, false)
				err := b.install()
				if err != nil {
					if strings.Contains(err.Error(), "ERR_PNPM_NO_MATCHING_VERSION") {
						return "", nil
					}
					return "", err
				}
				dts, err := b.resloveDTS(b.resolveEntry(dtsModule))
				if err != nil {
					return "", err
				}
				if dts != "" {
					// use tilde semver range instead of the exact version
					return strings.ReplaceAll(dts, fmt.Sprintf("%s@%s", typesPkgName, p.Version), fmt.Sprintf("%s@~%s", typesPkgName, p.Version)), nil
				}
			}
		}
	}

	return "", nil
}

func (ctx *BuildContext) normalizePackageJSON(p *PackageJSON) {
	if ctx.esmPath.GhPrefix || ctx.esmPath.PrPrefix {
		// if the name in package.json is not the same as the repository name
		if p.Name != ctx.esmPath.PkgName {
			p.PkgName = p.Name
			p.Name = ctx.esmPath.PkgName
		}
		p.Version = ctx.esmPath.PkgVersion
	} else {
		p.Version = strings.TrimPrefix(p.Version, "v")
	}

	if ctx.target == "types" {
		return
	}

	if p.Module == "" {
		if p.ES2015 != "" && ctx.existsPkgFile(p.ES2015) {
			p.Module = p.ES2015
		} else if p.JsNextMain != "" && ctx.existsPkgFile(p.JsNextMain) {
			p.Module = p.JsNextMain
		} else if p.Main != "" && (p.Type == "module" || strings.HasSuffix(p.Main, ".mjs")) {
			p.Module = p.Main
			p.Main = ""
		}
	}

	// Check if the `SubPath` is the same as the `main` or `module` field of the package.json
	if subModule := ctx.esmPath.SubBareName; subModule != "" {
		isPkgMainModule := false
		check := func(s string) bool {
			return isPkgMainModule || (s != "" && subModule == utils.NormalizePathname(stripModuleExt(s))[1:])
		}
		if p.Exports != nil {
			if s, ok := p.Exports.(string); ok {
				isPkgMainModule = check(s)
			} else if om, ok := p.Exports.(*OrderedMap); ok {
				if v := om.Get("."); v != nil {
					if s, ok := v.(string); ok {
						// exports: { ".": "./index.js" }
						isPkgMainModule = check(s)
					} else if om, ok := v.(*OrderedMap); ok {
						// exports: { ".": { "require": "./cjs/index.js", "import": "./esm/index.js" } }
						// exports: { ".": { "node": { "require": "./cjs/index.js", "import": "./esm/index.js" } } }
						// ...
						paths := getAllExportsPaths(om)
						for _, path := range paths {
							if check(path) {
								isPkgMainModule = true
								break
							}
						}
					}
				}
			}
		}
		if !isPkgMainModule {
			isPkgMainModule = (p.Module != "" && check(p.Module)) || (p.Main != "" && check(p.Main))
		}
		if isPkgMainModule {
			ctx.esmPath.SubBareName = ""
			ctx.esmPath.SubPath = ""
			ctx.path = ""
		}
	}
}

func (ctx *BuildContext) lexer(entry *BuildEntry, forceCjsOnly bool) (ret *BuildMeta, reexport string, err error) {
	if entry.esm != "" && !forceCjsOnly {
		if strings.HasSuffix(entry.esm, ".vue") || strings.HasSuffix(entry.esm, ".svelte") {
			ret = &BuildMeta{
				HasDefaultExport: true,
				NamedExports:     []string{"default"},
			}
			return
		}

		isESM, namedExports, e := ctx.esmLexer(entry.esm)
		if e != nil {
			err = e
			return
		}

		if isESM {
			ret = &BuildMeta{
				NamedExports:     namedExports,
				HasDefaultExport: contains(namedExports, "default"),
			}
			return
		}

		log.Warnf("fake ES module '%s' of '%s'", entry.esm, ctx.packageJson.Name)

		var r cjsModuleLexerResult
		r, err = ctx.cjsLexer(entry.esm)
		if err != nil {
			return
		}

		ret = &BuildMeta{
			HasDefaultExport: r.HasDefaultExport,
			NamedExports:     r.NamedExports,
			FromCJS:          true,
		}
		entry.cjs = entry.esm
		entry.esm = ""
		reexport = r.ReExport
		return
	}

	if entry.cjs != "" {
		var cjs cjsModuleLexerResult
		cjs, err = ctx.cjsLexer(entry.cjs)
		if err != nil {
			return
		}
		ret = &BuildMeta{
			HasDefaultExport: cjs.HasDefaultExport,
			NamedExports:     cjs.NamedExports,
			FromCJS:          true,
		}
		reexport = cjs.ReExport
	}
	return
}

func (ctx *BuildContext) cjsLexer(specifier string) (cjs cjsModuleLexerResult, err error) {
	cjs, err = cjsModuleLexer(ctx.npmrc, ctx.esmPath.PkgName, ctx.wd, specifier, ctx.getNodeEnv())
	if err == nil && cjs.Error != "" {
		err = fmt.Errorf("cjsLexer: %s", cjs.Error)
	}
	return
}

func (ctx *BuildContext) esmLexer(specifier string) (isESM bool, namedExports []string, err error) {
	isESM, namedExports, err = validateModuleFile(path.Join(ctx.wd, "node_modules", ctx.esmPath.PkgName, specifier))
	if err != nil {
		err = fmt.Errorf("esmLexer: %v", err)
	}
	return
}

func matchAsteriskExports(epxortsKey string, pkg ESMPath) (diff string, match bool) {
	if strings.ContainsRune(epxortsKey, '*') {
		prefix, _ := utils.SplitByLastByte(epxortsKey, '*')
		if subModule := "./" + pkg.SubBareName; strings.HasPrefix(subModule, prefix) {
			return strings.TrimPrefix(subModule, prefix), true
		}
	}
	return "", false
}

func resloveAsteriskPathMapping(om *OrderedMap, diff string) *OrderedMap {
	resovedConditions := newOrderedMap()
	for _, key := range om.keys {
		value := om.Get(key)
		if s, ok := value.(string); ok {
			resovedConditions.Set(key, strings.ReplaceAll(s, "*", diff))
		} else if om, ok := value.(*OrderedMap); ok {
			resovedConditions.Set(key, resloveAsteriskPathMapping(om, diff))
		}
	}
	return resovedConditions
}

func getAllExportsPaths(om *OrderedMap) []string {
	om.lock.RLock()
	defer om.lock.RUnlock()
	values := make([]string, 0, 5*len(om.keys))
	for _, key := range om.keys {
		v := om.m[key]
		if s, ok := v.(string); ok {
			values = append(values, s)
		} else if om2, ok := v.(*OrderedMap); ok {
			values = append(values, getAllExportsPaths(om2)...)
		}
	}
	return values
}

// make sure the entry is a relative specifier with extension
func normalizeBuildEntry(ctx *BuildContext, entry *BuildEntry) {
	if entry.esm != "" {
		entry.esm = normalizeEntryPath(entry.esm)
		if !endsWith(entry.esm, ".mjs", ".js") {
			if ctx.existsPkgFile(entry.esm + ".mjs") {
				entry.esm = entry.esm + ".mjs"
			} else if ctx.existsPkgFile(entry.esm + ".js") {
				entry.esm = entry.esm + ".js"
			} else if ctx.existsPkgFile(entry.esm, "index.mjs") {
				entry.esm = entry.esm + "/index.mjs"
			} else if ctx.existsPkgFile(entry.esm, "index.js") {
				entry.esm = entry.esm + "/index.js"
			}
		}
	}

	if entry.cjs != "" {
		entry.cjs = normalizeEntryPath(entry.cjs)
		if !endsWith(entry.cjs, ".cjs", ".js") {
			if ctx.existsPkgFile(entry.cjs + ".cjs") {
				entry.cjs = entry.cjs + ".cjs"
			} else if ctx.existsPkgFile(entry.cjs + ".js") {
				entry.cjs = entry.cjs + ".js"
			} else if ctx.existsPkgFile(entry.cjs, "index.cjs") {
				entry.cjs = entry.cjs + "/index.cjs"
			} else if ctx.existsPkgFile(entry.cjs, "index.js") {
				entry.cjs = entry.cjs + "/index.js"
			}
		}
		// check if the cjs entry is an ESM
		if entry.cjs != "" && strings.HasSuffix(entry.cjs, ".js") {
			isESM, _, _ := validateModuleFile(path.Join(ctx.pkgDir, entry.cjs))
			if isESM {
				if entry.esm == "" {
					entry.esm = entry.cjs
				}
				entry.cjs = ""
			}
		}
	}

	if entry.esm != "" {
		entry.esm = normalizeEntryPath(entry.esm)
	}
	if entry.cjs != "" {
		entry.cjs = normalizeEntryPath(entry.cjs)
	}
	if entry.dts != "" {
		entry.dts = normalizeEntryPath(entry.dts)
	}
}

func normalizeEntryPath(path string) string {
	return "." + utils.NormalizePathname(path)
}

func normalizeSavePath(zoneId string, pathname string) string {
	segs := strings.Split(pathname, "/")
	for i, seg := range segs {
		if strings.HasPrefix(seg, "X-") && len(seg) > 42 {
			h := sha1.New()
			h.Write([]byte(seg))
			segs[i] = "X-" + hex.EncodeToString(h.Sum(nil))
		}
	}
	if zoneId != "" {
		return zoneId + "/" + strings.Join(segs, "/")
	}
	return strings.Join(segs, "/")
}
