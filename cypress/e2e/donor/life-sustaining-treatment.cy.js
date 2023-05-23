describe('Life sustaining treatment', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/life-sustaining-treatment&withDonorDetails=1&withAttorney=1&withType=hw');
    });

    it('can be agreed to', () => {
        cy.checkA11yApp();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/restrictions');
        cy.contains('life-sustaining treatment');
    });

    it('can be disagreed with', () => {
        cy.checkA11yApp();

        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/restrictions');
        cy.contains('life-sustaining treatment').should('not.exist');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select if you do or do not give your attorneys authority to give or refuse consent to life-sustaining treatment on your behalf');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select if you do or do not give your attorneys authority to give or refuse consent to life-sustaining treatment on your behalf');
    });
});
