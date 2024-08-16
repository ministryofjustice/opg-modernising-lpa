describe('Verify donor details', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/verify-donor-details&progress=confirmYourName');
    });

    it('can confirm the details', () => {
        cy.checkA11yApp();

        cy.contains('dl', 'Sam');
        cy.contains('dl', 'Smith');
        cy.contains('dl', '2 January 2000');
        cy.contains('dl', '1 RICHMOND PLACE');

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-task-list li:nth-child(2)').should('contain', 'Completed');
        cy.get('.govuk-task-list li:nth-child(2) a').should('not.exist');
    });
});
