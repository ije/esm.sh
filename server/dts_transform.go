package server

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/esm-dev/esm.sh/server/storage"
	"github.com/ije/gox/set"
	"github.com/ije/gox/utils"
)

func (ctx *BuildContext) transformDTS(dts string) error {
	start := time.Now()
	n, err := transformDTS(ctx, dts, ctx.getBuildArgsPrefix(true), nil)
	if err != nil {
		return err
	}
	if DEBUG {
		log.Debugf("transform dts '%s'(%d related dts files) in %v", dts, n, time.Since(start))
	}
	return nil
}

// transformDTS transforms a `.d.ts` file for deno/editor-lsp
func transformDTS(ctx *BuildContext, dts string, buildArgsPrefix string, marker *set.Set[string]) (n int, err error) {
	isEntry := marker == nil
	if isEntry {
		marker = set.New[string]()
	}
	dtsPath := path.Join("/"+ctx.esmPath.Name(), buildArgsPrefix, dts)
	if marker.Has(dtsPath) {
		// don't transform repeatly
		return
	}
	marker.Add(dtsPath)

	savePath := normalizeSavePath(ctx.zoneId, path.Join("types", dtsPath))
	// check if the dts file has been transformed
	_, err = buildStorage.Stat(savePath)
	if err == nil || err != storage.ErrNotFound {
		return
	}

	dtsFilePath := path.Join(ctx.wd, "node_modules", ctx.esmPath.PkgName, dts)
	dtsWd := path.Dir(dtsFilePath)
	dtsFile, err := os.Open(dtsFilePath)
	if err != nil {
		// if the dts file does not exist, print a warning but continue to build
		if os.IsNotExist(err) {
			if isEntry {
				err = fmt.Errorf("types not found")
			} else {
				log.Warnf("dts not found: %s", dtsFilePath)
				err = nil
			}
		}
		return
	}

	buffer, recycle := NewBuffer()
	defer recycle()

	internalDts := set.New[string]()

	err = parseDts(dtsFile, buffer, func(specifier string, kind TsImportKind, position int) (string, error) {
		if ctx.esmPath.PkgName == "@types/node" {
			if strings.HasPrefix(specifier, "node:") || nodeBuiltinModules[specifier] || isRelPathSpecifier(specifier) {
				return specifier, nil
			}
		}

		// normalize specifier
		specifier = normalizeImportSpecifier(specifier)

		if isRelPathSpecifier(specifier) {
			specifier = strings.TrimSuffix(specifier, ".d")
			if !endsWith(specifier, ".d.ts", ".d.mts") {
				var p PackageJSONRaw
				var hasTypes bool
				if utils.ParseJSONFile(path.Join(dtsWd, specifier, "package.json"), &p) == nil {
					dir := path.Join("/", path.Dir(dts))
					if types := p.Types.MainString(); types != "" {
						specifier, _ = relPath(dir, "/"+path.Join(dir, specifier, types))
						hasTypes = true
					} else if typings := p.Typings.MainString(); typings != "" {
						specifier, _ = relPath(dir, "/"+path.Join(dir, specifier, typings))
						hasTypes = true
					}
				}
				if !hasTypes {
					if existsFile(path.Join(dtsWd, specifier+".d.mts")) {
						specifier = specifier + ".d.mts"
					} else if existsFile(path.Join(dtsWd, specifier+".d.ts")) {
						specifier = specifier + ".d.ts"
					} else if existsFile(path.Join(dtsWd, specifier, "index.d.mts")) {
						specifier = strings.TrimSuffix(specifier, "/") + "/index.d.mts"
					} else if existsFile(path.Join(dtsWd, specifier, "index.d.ts")) {
						specifier = strings.TrimSuffix(specifier, "/") + "/index.d.ts"
					} else if endsWith(specifier, ".js", ".mjs", ".cjs") {
						specifier = stripModuleExt(specifier)
						if existsFile(path.Join(dtsWd, specifier+".d.mts")) {
							specifier = specifier + ".d.mts"
						} else if existsFile(path.Join(dtsWd, specifier+".d.ts")) {
							specifier = specifier + ".d.ts"
						}
					}
				}
			}

			if endsWith(specifier, ".d.ts", ".d.mts") {
				internalDts.Add(specifier)
			} else {
				specifier += ".d.ts"
			}
			return specifier, nil
		}

		if kind == TsReferenceTypes && specifier == "node" {
			// return empty string to ignore the reference types 'node'
			return "", nil
		}

		if specifier == "node" || isNodeBuiltInModule(specifier) {
			return specifier, nil
		}

		depPkgName, _, subPath, _ := splitEsmPath(specifier)
		specifier = depPkgName
		if subPath != "" {
			specifier += "/" + subPath
		}

		if depPkgName == ctx.esmPath.PkgName {
			if strings.ContainsRune(subPath, '*') {
				return fmt.Sprintf(
					"{ESM_CDN_ORIGIN}/%s/%s%s",
					ctx.esmPath.Name(),
					ctx.getBuildArgsPrefix(true),
					subPath,
				), nil
			} else {
				entry := ctx.resolveEntry(EsmPath{
					PkgName:       depPkgName,
					PkgVersion:    ctx.esmPath.PkgVersion,
					SubPath:       subPath,
					SubModuleName: subPath,
				})
				if entry.types != "" {
					return fmt.Sprintf(
						"{ESM_CDN_ORIGIN}/%s/%s%s",
						ctx.esmPath.Name(),
						ctx.getBuildArgsPrefix(true),
						strings.TrimPrefix(entry.types, "./"),
					), nil
				}
			}
			return specifier, nil
		}

		// respect `?alias` query
		alias, ok := ctx.args.alias[depPkgName]
		if ok {
			aliasPkgName, _, aliasSubPath, _ := splitEsmPath(alias)
			depPkgName = aliasPkgName
			if aliasSubPath != "" {
				if subPath != "" {
					subPath = aliasSubPath + "/" + subPath
				} else {
					subPath = aliasSubPath
				}
			}
			specifier = depPkgName
			if subPath != "" {
				specifier += "/" + subPath
			}
		}

		// respect `?external` query
		if ctx.externalAll || ctx.args.external.Has(depPkgName) {
			return specifier, nil
		}

		typesPkgName := toTypesPackageName(depPkgName)
		if _, ok := ctx.pkgJson.Dependencies[typesPkgName]; ok {
			depPkgName = typesPkgName
		} else if _, ok := ctx.pkgJson.PeerDependencies[typesPkgName]; ok {
			depPkgName = typesPkgName
		}

		_, p, err := ctx.lookupDep(depPkgName, true)
		if err != nil {
			if kind == TsDeclareModule && strings.HasSuffix(err.Error(), " not found") {
				return specifier, nil
			}
			return "", err
		}

		dtsModule := EsmPath{
			PkgName:       p.Name,
			PkgVersion:    p.Version,
			SubPath:       subPath,
			SubModuleName: subPath,
		}
		args := BuildArgs{}
		b := &BuildContext{
			esmPath: dtsModule,
			npmrc:   ctx.npmrc,
			args:    args,
			target:  "types",
			zoneId:  ctx.zoneId,
		}
		err = b.install()
		if err != nil {
			return "", err
		}

		dtsPath, err := b.resloveDTS(b.resolveEntry(dtsModule))
		if err != nil {
			return "", err
		}

		if dtsPath != "" {
			return fmt.Sprintf("{ESM_CDN_ORIGIN}%s", dtsPath), nil
		}

		if kind == TsDeclareModule {
			return fmt.Sprintf("{ESM_CDN_ORIGIN}/%s", dtsModule.Specifier()), nil
		}

		return fmt.Sprintf("{ESM_CDN_ORIGIN}%s", b.Path()), nil
	})
	if err != nil {
		return
	}
	dtsFile.Close()

	err = buildStorage.Put(savePath, bytes.NewReader(ctx.rewriteDTS(dts, buffer.Bytes())))
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	var errors []error
	for _, s := range internalDts.Values() {
		wg.Add(1)
		go func(s string) {
			j, err := transformDTS(ctx, "./"+path.Join(path.Dir(dts), s), buildArgsPrefix, marker)
			if err != nil {
				errors = append(errors, err)
			}
			n += j
			wg.Done()
		}(s)
	}
	wg.Wait()

	if len(errors) > 0 {
		err = errors[0]
	}
	return
}
