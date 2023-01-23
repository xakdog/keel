// See https://vitest.dev/guide/extending-matchers.html for docs
// on typing custom matchers

interface CustomMatchers<R = unknown> {
  toHaveAuthorizationError(): R;
}

declare global {
  namespace Vi {
    interface Assertion extends CustomMatchers {}
    interface AsymmetricMatchersContaining extends CustomMatchers {}
  }
}

export {};
