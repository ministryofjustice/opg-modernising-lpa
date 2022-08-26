describe('Sign in using GOV UK Sign In service', () => {
    context('accessing home page without logging in', () => {
        it('does not show user email', () => {
            cy.visit('/home');
            cy.contains('h1', 'User not signed in');
        });
    });

    context('with an existing GOV UK account', () => {
        it('can authenticate with cy.request', () => {
            cy.visit('/start');
            cy.contains('a', 'Start').click();

            cy.url().should('contain', '/home');
            cy.contains('Welcome gideon.felix@example.org');
        });
    });
})
