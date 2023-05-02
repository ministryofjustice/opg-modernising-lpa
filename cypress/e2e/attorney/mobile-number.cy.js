import {TestMobile} from "../../support/e2e";

describe('Mobile number', () => {
    beforeEach(() => {
        cy.visit('/testing-start?withAttorney=1&asAttorney=1&redirect=/attorney-mobile-number');
    });

    it('can be completed', () => {
        cy.checkA11yApp();

        cy.get('#f-mobile').type(TestMobile);

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/attorney-sign');
    });

    it('can be empty', () => {
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/attorney-sign');
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
