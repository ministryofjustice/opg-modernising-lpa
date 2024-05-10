const { TOTP } = require("totp-generator");

describe('Smoke tests', () => {
  describe('external dependencies', () => {
    describe('UID service', () => {
      it('request signing and base URL are configured correctly', () => {
        cy.request('/health-check/dependency').should((response) => {
          expect(response.status).not.to.eq(403)
        })
      })
    })
  })

  describe('app', () => {
    it('is available', { requestTimeout: 30000, defaultCommandTimeout: 30000 }, () => {
      cy.visit('/')

      cy.get('h1').should('contain', 'Make and register a lasting power of attorney (LPA)');

      if (Cypress.config().baseUrl.includes('preproduction')) {
        cy.intercept('https://signin.integration.account.gov.uk/**', (req) => {
          req.headers['Authorization'] = 'Basic ' + btoa(Cypress.env('TEST_ONELOGIN_BASIC_AUTH'));
        });

        cy.contains('a', 'Start').click({ timeout: 30000 });
        cy.wait(10000);

        const { otp } = TOTP.generate(Cypress.env('TEST_ONELOGIN_TOTP_KEY'));

        cy.origin('https://signin.integration.account.gov.uk', { args: { token: otp } }, ({ token }) => {
          cy.url().should('contain', '/sign-in-or-create');

          cy.contains('Sign in').click();
          cy.get('[type=email]').type('opgteam+modernising-lpa@digital.justice.gov.uk');
          cy.get('form').submit();
          cy.get('[type=password]').type(Cypress.env('TEST_ONELOGIN_PASSWORD'), { parseSpecialCharSequences: false });
          cy.get('form').submit();

          cy.get('[name=code]').type(token);
          cy.contains('button', 'Continue').click();
          cy.wait(10000);
        });

        cy.url().should('contain', '/dashboard');
        cy.contains('Manage your LPAs');
      } else {
        cy.contains('a', 'Start');
      }
    })
  })
})
