/// <reference types='codeceptjs' />
type steps_file = typeof import('./codecept/steps_file.js');
type AxeRunner = import('./codecept/helpers/axeRunner_helper.js');

declare namespace CodeceptJS {
  interface SupportObject { I: I, current: any }
  interface Methods extends Playwright, AxeRunner {}
  interface I extends ReturnType<steps_file>, WithTranslation<AxeRunner> {}
  namespace Translation {
    interface Actions {}
  }
}
