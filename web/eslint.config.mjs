import tsPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import eslintConfigPrettier from "eslint-config-prettier";
import pluginPrettier from "eslint-plugin-prettier";

const tsRecommended = tsPlugin.configs["flat/recommended"].map((config) => {
  if (!config.languageOptions) {
    return config;
  }

  return {
    ...config,
    languageOptions: {
      ...config.languageOptions,
      parser: tsParser,
    },
  };
});

export default [
  {
    ignores: ["node_modules/**", "static/**", "tmp/**"],
  },
  ...tsRecommended,
  eslintConfigPrettier,
  {
    files: ["assets/**/*.ts", "global.d.ts"],
    languageOptions: {
      parser: tsParser,
      sourceType: "module",
      parserOptions: {
        ecmaVersion: "latest",
        project: "./tsconfig.json",
        tsconfigRootDir: import.meta.dirname,
      },
      globals: {
        window: "readonly",
        document: "readonly",
        Alpine: "readonly",
      },
    },
    plugins: {
      prettier: pluginPrettier,
    },
    rules: {
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "prettier/prettier": "error",
    },
  },
];
