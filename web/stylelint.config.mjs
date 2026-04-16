export default {
  extends: [
    'stylelint-config-standard',
    'stylelint-config-standard-scss',
    'stylelint-config-recommended-vue/scss',
  ],
  overrides: [
    { files: ['**/*.vue'], customSyntax: 'postcss-html' },
    { files: ['**/*.scss'], customSyntax: 'postcss-scss' },
  ],
  rules: {
    'no-descending-specificity': null,
    'no-empty-source': null,
    // BEM convention (__element, --modifier) + allow uppercase for third-party class targeting
    'selector-class-pattern': null,
    // Preserve author-intended blank lines (used as section dividers in theme token files)
    'declaration-empty-line-before': null,
    'custom-property-empty-line-before': null,
    // Keep `currentColor` camelCase; both spellings are spec-valid
    'value-keyword-case': null,
    'selector-pseudo-class-no-unknown': [
      true,
      { ignorePseudoClasses: ['deep', 'slotted', 'global'] },
    ],
    'selector-pseudo-element-no-unknown': [
      true,
      { ignorePseudoElements: ['v-deep', 'v-slotted', 'v-global'] },
    ],
    'at-rule-no-unknown': [
      true,
      {
        ignoreAtRules: [
          'apply',
          'screen',
          'variants',
          'tailwind',
          'use',
          'forward',
          'mixin',
          'include',
          'if',
          'else',
          'each',
          'for',
          'function',
          'return',
        ],
      },
    ],
  },
  ignoreFiles: ['dist/**', 'node_modules/**', 'public/**', 'src/**/*.d.ts'],
}
