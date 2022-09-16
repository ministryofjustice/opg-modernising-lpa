const { defineConfig } = require("cypress");

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://localhost:5050',
    experimentalSessionAndOrigin: true,
    pageLoadTimeout: 5000,
    setupNodeEvents(on, config) {
      on('task', {
        log(message) {
          console.log(message)

          return null
        },
      })
    },
    video: false
  },
});
