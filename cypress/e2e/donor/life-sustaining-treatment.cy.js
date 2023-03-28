describe('Life sustaining treatment', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/life-sustaining-treatment&withDonorDetails=1&withAttorney=1');
    });

    it('can be submitted', () => {
        cy.checkA11yApp();

        cy.contains('label', 'Option A').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/restrictions');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select if the donor gives or does not give their attorneys authority to consent to life-sustainng treatment');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select if the donor gives or does not give their attorneys authority to consent to life-sustainng treatment');
    });
});
