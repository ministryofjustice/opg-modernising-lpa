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
});
