describe('Who do you want to be certificate provider guidance', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/who-do-you-want-to-be-certificate-provider-guidance');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });
});
