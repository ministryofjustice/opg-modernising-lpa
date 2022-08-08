const { setHeadlessWhen, setCommonPlugins } = require('@codeceptjs/configure');

// turn on headless mode when running with HEADLESS=true environment variable
// export HEADLESS=true && npx codeceptjs run
setHeadlessWhen(process.env.HEADLESS);

// enable all common plugins https://github.com/codeceptjs/configure#setcommonplugins
setCommonPlugins();

exports.config = {
  tests: './codecept/scenarios/*_test.js',
  output: './codecept/scenarios/output',
  helpers: {
    Playwright: {
      url: 'http://localhost:5050',
      show: false,
      browser: 'chromium'
    },
    AxeRunner: {
      require: './codecept/helpers/axeRunner_helper.js'
    }
  },
  include: {
    I: './codecept/steps_file.js'
  },
  bootstrap: null,
  mocha: {},
  name: 'app'
}
