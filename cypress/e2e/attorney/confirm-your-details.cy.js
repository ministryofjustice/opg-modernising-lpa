import { TestMobile } from '../../support/e2e';

describe('Confirm your details', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/mobile-number&lpa.attorneys=1&attorneyProvided=1&loginAs=attorney');

        cy.get('#f-mobile').type(TestMobile);
        cy.contains('Continue').click();
    });

    it('shows details', () => {
        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('2 January 2000');
        cy.contains('John Smith');
        cy.contains('2 RICHMOND PLACE');
        cy.contains('07700900000');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/read-the-lpa');
    });
});
