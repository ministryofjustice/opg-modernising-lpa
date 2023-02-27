describe('How you know donor', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-do-you-know-the-donor&completeLpa=1&asCertificateProvider=1');
    });

    describe('when personally', () => {
        it('can be completed', () => {
            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });
            
            cy.contains('label', 'Personally').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/how-long-have-you-known-donor');
            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.contains('label', '2 years or more').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/certificate-provider-your-details');
        });
    });

    describe('when professionally', () => {
        it('can be completed', () => {
            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });
            
            cy.contains('label', 'In a professional capacity').click();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/certificate-provider-your-details');
        });
    });
});
