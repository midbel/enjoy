package eval

import (
  "github.com/midbel/enjoy/ast"
  "github.com/midbel/enjoy/env"
  "github.com/midbel/enjoy/value"
)

func evalImport(i ast.ImportNode, ev env.Environ[value.Value]) (value.Value, error) {
  return nil, nil
}

func evalExport(i ast.ExportNode, ev env.Environ[value.Value]) (value.Value, error) {
  return nil, nil
}

func evalExportFrom(i ast.ExportFromNode, ev env.Environ[value.Value]) (value.Value, error) {
  return nil, nil
}
