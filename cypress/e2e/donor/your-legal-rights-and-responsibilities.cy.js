describe('Your legal rights and responsibilities', () => {
    describe('when signed out', () => {
        it('is accessible from the footer', () => {
            cy.visit('/');

            cy.contains('footer a', 'Your legal rights and responsibilities').click();
            cy.url().should('contain', '/your-legal-rights-and-responsibilities');

            cy.contains('a', 'Continue to signing page').should('not.exist');
        });
    });

    describe('when signed in', () => {
        it('is accessible from the footer', () => {
            cy.visit('/fixtures?redirect=/your-legal-rights-and-responsibilities&progress=confirmYourIdentity');
            cy.contains('a', 'Continue to signing page');

            cy.contains('footer a', 'Your legal rights and responsibilities').click();
            cy.url().should('contain', '/your-legal-rights-and-responsibilities');

            cy.contains('a', 'Continue to signing page').should('not.exist');
        });
    });
});
