const { defineConfig } = require("cypress");

module.exports = defineConfig({
  projectId: "xxbft5",
  e2e: {
    baseUrl: 'http://localhost:5050',
    experimentalSessionAndOrigin: true,
    defaultCommandTimeout: 2000,
    pageLoadTimeout: 3000,
    setupNodeEvents(on, config) {
      on('task', {
        log(message) {
          console.log(message)

          return null
        },
        table(message) {
          console.table(message)

          return null
        }
      })
    },
    video: false
  },
});
