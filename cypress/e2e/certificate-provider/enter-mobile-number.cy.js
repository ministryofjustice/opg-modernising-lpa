import {TestMobile} from "../../support/e2e";

describe('Enter mobile number', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-enter-mobile-number&asCertificateProvider=1');
    });

    it('can be completed', () => {
        cy.checkA11yApp();

        cy.get('#f-mobile').type(TestMobile);

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-your-address');
    });

    it('errors when empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter mobile number');
        });

        cy.contains('[for=f-mobile] ~ .govuk-error-message', 'Enter mobile number');
    });

    it('errors when not a UK mobile', () => {
        cy.get('#f-mobile').type('not a mobile');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Mobile number must be a UK mobile number, like 07700 900 982 or +44 7700 900 982');
        });

        cy.contains('[for=f-mobile] ~ .govuk-error-message', 'Mobile number must be a UK mobile number, like 07700 900 982 or +44 7700 900 982');
    });
});
