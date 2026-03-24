import fs from 'fs';
import { Project, SyntaxKind } from 'ts-morph';

interface PatchArgs {
  configPath: string;
  imports: string[];
  plugins: string[];
  spreads: string[];
  jsonMerge: Record<string, unknown[]> | null;
}

function parseArgs(): PatchArgs {
  const args = process.argv.slice(2);
  const configPath = args[0];

  if (!configPath) {
    console.error('Usage: patcher <config-path> [flags]');
    process.exit(1);
  }

  const getFlag = (name: string): string[] => {
    const idx = args.indexOf(`--${name}`);
    if (idx === -1 || idx + 1 >= args.length) return [];
    return JSON.parse(args[idx + 1]);
  };

  const getRawFlag = (
    name: string,
  ): Record<string, unknown[]> | null => {
    const idx = args.indexOf(`--${name}`);
    if (idx === -1 || idx + 1 >= args.length) return null;
    return JSON.parse(args[idx + 1]);
  };

  return {
    configPath,
    imports: getFlag('imports'),
    plugins: getFlag('plugins'),
    spreads: getFlag('spreads'),
    jsonMerge: getRawFlag('json-merge'),
  };
}

// JSON patching: merge arrays into existing JSON object
function patchJson(
  configPath: string,
  merge: Record<string, unknown[]>,
): void {
  const raw = fs.readFileSync(configPath, 'utf-8');
  const obj = JSON.parse(raw);

  for (const [key, values] of Object.entries(merge)) {
    if (!Array.isArray(obj[key])) {
      obj[key] = [];
    }
    obj[key].push(...values);
  }

  fs.writeFileSync(
    configPath,
    JSON.stringify(obj, null, 2) + '\n',
  );
  console.log('Patched ' + configPath);
}

// JS/TS patching: imports, plugins, spreads
function patchJs(opts: PatchArgs): void {
  const project = new Project();
  const sourceFile = project.addSourceFileAtPath(opts.configPath);

  // Add import statements after existing imports
  if (opts.imports.length > 0) {
    const stmts = sourceFile.getStatements();
    let lastImportIdx = -1;
    for (let i = 0; i < stmts.length; i++) {
      if (stmts[i].getKind() === SyntaxKind.ImportDeclaration) {
        lastImportIdx = i;
      }
    }

    for (let i = 0; i < opts.imports.length; i++) {
      sourceFile.insertStatements(
        lastImportIdx + 1 + i,
        opts.imports[i],
      );
    }
  }

  // Add plugins to defineConfig's vite.plugins array
  if (opts.plugins.length > 0) {
    const callExprs = sourceFile.getDescendantsOfKind(
      SyntaxKind.CallExpression,
    );
    const defineConfigCall = callExprs.find(
      (c) => c.getExpression().getText() === 'defineConfig',
    );

    if (!defineConfigCall) {
      console.error(
        'defineConfig() not found in ' + opts.configPath,
      );
      process.exit(1);
    }

    const configArgs = defineConfigCall.getArguments();
    let configObj;

    if (
      configArgs.length === 0 ||
      configArgs[0].getKind() !==
        SyntaxKind.ObjectLiteralExpression
    ) {
      if (configArgs.length > 0) {
        configArgs[0].replaceWithText('{}');
      } else {
        defineConfigCall.addArgument('{}');
      }
      configObj = defineConfigCall
        .getArguments()[0]
        .asKindOrThrow(SyntaxKind.ObjectLiteralExpression);
    } else {
      configObj = configArgs[0].asKindOrThrow(
        SyntaxKind.ObjectLiteralExpression,
      );
    }

    let viteProp = configObj.getProperty('vite');
    if (!viteProp) {
      configObj.addPropertyAssignment({
        name: 'vite',
        initializer: '{\n    plugins: []\n  }',
      });
      viteProp = configObj.getProperty('vite')!;
    }

    const viteObj = viteProp
      .asKindOrThrow(SyntaxKind.PropertyAssignment)
      .getInitializerIfKindOrThrow(
        SyntaxKind.ObjectLiteralExpression,
      );

    let pluginsProp = viteObj.getProperty('plugins');
    if (!pluginsProp) {
      viteObj.addPropertyAssignment({
        name: 'plugins',
        initializer: '[]',
      });
      pluginsProp = viteObj.getProperty('plugins')!;
    }

    const pluginsArr = pluginsProp
      .asKindOrThrow(SyntaxKind.PropertyAssignment)
      .getInitializerIfKindOrThrow(
        SyntaxKind.ArrayLiteralExpression,
      );

    for (const plugin of opts.plugins) {
      pluginsArr.addElement(plugin);
    }
  }

  // Add spreads to default export array
  if (opts.spreads.length > 0) {
    const exportAssignment = sourceFile
      .getStatements()
      .find(
        (s) =>
          s.getKind() === SyntaxKind.ExportAssignment,
      );

    if (!exportAssignment) {
      console.error(
        'Default export not found in ' + opts.configPath,
      );
      process.exit(1);
    }

    const exportArr = exportAssignment
      .getDescendantsOfKind(SyntaxKind.ArrayLiteralExpression)[0];

    if (!exportArr) {
      console.error(
        'Default export is not an array in ' + opts.configPath,
      );
      process.exit(1);
    }

    for (const spread of opts.spreads) {
      exportArr.addElement(spread);
    }
  }

  sourceFile.saveSync();
  console.log('Patched ' + opts.configPath);
}

function main(): void {
  const opts = parseArgs();
  const ext = opts.configPath.split('.').pop() ?? '';
  const basename = opts.configPath.split('/').pop() ?? '';

  // Detect JSON files by extension or dotfile pattern
  const isJson =
    ext === 'json' ||
    (basename.startsWith('.') &&
      !['js', 'mjs', 'ts', 'mts'].includes(ext));

  if (isJson && opts.jsonMerge) {
    patchJson(opts.configPath, opts.jsonMerge);
  } else if (!isJson) {
    patchJs(opts);
  }
}

main();
