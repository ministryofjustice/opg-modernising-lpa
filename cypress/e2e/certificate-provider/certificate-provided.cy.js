describe('Certificate provided', () => {
    it('has a print this page link', () => {
        cy.on('window:before:load', (win) => {
            cy.stub(win, 'print')
        })

        cy.visit('/fixtures/certificate-provider?redirect=/certificate-provided&progress=confirmYourIdentity')

        cy.checkA11yApp();

        cy.contains('a', 'Print this page').click();
        cy.window().its('print').should('be.called')
    });

    it('has a button to the dashboard', () => {
        cy.visit('/fixtures/certificate-provider?redirect=/certificate-provided&progress=confirmYourIdentity')

        cy.contains('a', 'Go to your dashboard').click();
        cy.url().should('contain', '/dashboard');
    });
})
