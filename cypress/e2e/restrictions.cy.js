describe('Restrictions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/restrictions&withAttorney=1');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-restrictions').type('this that');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/who-do-you-want-to-be-certificate-provider-guidance');
    });

    it('errors when too long', () => {
        cy.get('#f-restrictions').invoke('val', 'a'.repeat(10001));
        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Restrictions and conditions must be 10000 characters or less');
        });
        
        cy.contains('[for=f-restrictions] + * +  .govuk-error-message', 'Restrictions and conditions must be 10000 characters or less');
    });
});
