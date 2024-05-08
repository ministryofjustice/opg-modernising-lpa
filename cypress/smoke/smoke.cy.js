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
    it('is available', () => {
      cy.visit('/')

      cy.get('h1').should('contain', 'Make and register a lasting power of attorney (LPA)');

      if (Cypress.config().baseUrl.includes('1221mlpab18')) {
        cy.intercept('https://signin.integration.account.gov.uk/**', (req) => {
          req.headers['Authorization'] = 'Basic aW50ZWdyYXRpb24tdXNlcjp3aW50ZXIyMDIx';
        });

        cy.contains('a', 'Start').click();

        const { otp } = TOTP.generate(Cypress.env('TEST_ONELOGIN_TOTP_KEY'));

        cy.origin('https://signin.integration.account.gov.uk', { args: { token: otp } }, ({ token }) => {
          cy.url().should('contain', '/sign-in-or-create')

          cy.contains('Sign in').click();
          cy.get('[type=email]').type('opgteam+modernising-lpa@digital.justice.gov.uk');
          cy.get('form').submit();
          cy.get('[type=password]').type(Cypress.env('TEST_ONELOGIN_PASSWORD'));
          cy.get('form').submit();

          cy.get('[name=code]').type(token);
          cy.contains('button', 'Continue').click();
        });

        cy.origin('https://preproduction.app.modernising.opg.service.justice.gov.uk', () => {
          cy.url().should('contain', '/dashboard');
          cy.contains('Manage your LPAs');
        });
      } else {
        cy.contains('a', 'Start');
      }
    })
  })
})
