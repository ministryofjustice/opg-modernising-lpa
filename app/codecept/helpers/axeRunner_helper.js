const Helper = require('@codeceptjs/helper');
const { injectAxe, checkA11y } = require('axe-playwright')

class AxeRunner extends Helper {

  // before/after hooks
  /**
   * @protected
   */
  _before() {

  }

  /**
   * @protected
   */
  _after() {
    // remove if not used
  }

  async runAccessibilityChecks() {
    const { page } = this.helpers.Playwright;
    await injectAxe(page)
    await checkA11y(page, null, {
      detailedReport: true,
    })
  }
  // add custom methods here
  // If you need to access other helpers
  // use: this.helpers['helperName']

}

module.exports = AxeRunner;
