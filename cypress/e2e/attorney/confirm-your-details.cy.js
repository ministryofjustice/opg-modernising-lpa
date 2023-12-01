import { TestMobile } from '../../support/e2e';

describe('Confirm your details', () => {
    beforeEach(() => {
        cy.visit('/fixtures/attorney?redirect=/mobile-number');

        cy.get('#f-mobile').type(TestMobile);
        cy.contains('Continue').click();

        cy.get('[name="language-preference"]').check('cy')
        cy.contains('button', 'Save and continue').click()
    });

    it('shows details', () => {
        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('2 January 2000');
        cy.contains('Jessie Jones');
        cy.contains('2 RICHMOND PLACE');
        cy.contains('07700900000');
        cy.contains('Welsh');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/read-the-lpa');
    });
});
