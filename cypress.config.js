import {defineConfig} from 'cypress';
import {plugin as cypressGrepPlugin} from '@cypress/grep/plugin'

export default defineConfig({
  projectId: "xxbft5",
  retries: {
    runMode: 1,
  },
  e2e: {
    baseUrl: 'http://localhost:5050',
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

      cypressGrepPlugin(config)

      return config
    },
    specPattern: '**/*.cy.js',
    video: false
  }
});
