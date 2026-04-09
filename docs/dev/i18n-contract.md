# Ech0 i18n Contract

## 1. Locale Strategy

- Supported locales: `zh-CN`, `en-US`.
- Default locale: `zh-CN`.
- Fallback locale: `en-US`.
- Negotiation order:
  1. Explicit request value (`lang` query or `X-Locale` header)
  2. User preference (`user.locale`, for authenticated requests)
  3. System default locale (`system_settings.default_locale`)
  4. `Accept-Language`
  5. Fallback locale (`zh-CN`)
- Frontend startup order:
  1. `localStorage.locale`
  2. user profile locale (after login)
  3. `system_settings.default_locale`
  4. browser locale

## 2. API Contract

- API responses keep `msg` for backward compatibility.
- For i18n-aware clients, responses may include:
  - `error_code`: stable machine-readable error code.
  - `message_key`: translation key.
  - `message_params`: interpolation parameters.
- Preferred client rendering order:
  1. `message_key` + `message_params`
  2. `msg`
  3. local fallback text

## 3. Key Naming Rules

- Use semantic keys, never use source text as key.
- Recommended format: `domain.module.action`.
- Current examples:
  - `common.request_failed`
  - `auth.token_missing`
  - `dashboard.logs.tail_invalid`

## 4. Backend Rules

- New business errors should use stable `error_code`.
- If the message is user-facing, attach `message_key`.
- Keep logs and telemetry machine-oriented; do not localize log field names.

## 5. Frontend Rules

- New UI text should be added to locale JSON and rendered via `t()`.
- API error display should prefer `message_key` from server.
- Send `X-Locale` on API requests to allow server-side localization.
- User settings should persist personal locale to backend (`user.locale`) and switch UI locale immediately after save.

## 6. Content/Template Rules

- User-generated content is not auto-translated by backend.
- System-generated content (notification templates, built-in prompts) should come from locale resources keyed by `message_key`.
