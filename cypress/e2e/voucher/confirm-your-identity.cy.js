describe('Confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/confirm-your-identity&progress=verifyDonorDetails');
    });

    it('can be confirmed', () => {
        cy.checkA11yApp();
        cy.contains('a', 'Continue').click();
        cy.contains('label', 'Vivian Vaughn').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/one-login-identity-details');
        cy.checkA11yApp();
        cy.contains('a', 'Continue').click();

        cy.get('.govuk-task-list li:nth-child(3)').should('contain', 'Completed');
        cy.contains('a', 'Confirm your identity').click();

        cy.url().should('contain', '/one-login-identity-details');
        cy.contains('a', 'Continue').click();

        cy.contains('a', 'Confirm your name').click();
        cy.contains('a', 'Change').should('not.exist');

        cy.contains('a', 'Manage your LPAs').click();
        cy.contains('I’m vouching for someone');
    });

    it('can fail', () => {
        cy.contains('a', 'Continue').click();
        cy.contains('label', 'Sam Smith').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/unable-to-confirm-identity');

        cy.contains('a', 'Manage your LPAs').click();
        cy.contains('I’m vouching for someone').should('not.exist');;
    });
});
