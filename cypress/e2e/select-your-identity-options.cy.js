describe('Select your identity options', () => {
    beforeEach(() => {
        cy.visit('/testing-start?withPayment=1&redirect=/select-your-identity-options');
    });

    it('can select on first page', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/your-chosen-identity-options');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Your GOV.UK One Login Identity');
        cy.contains('button', 'Continue');
    });

    it('can select on second page', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/select-your-identity-options-1');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'Your passport').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/your-chosen-identity-options');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('passport');        
        cy.contains('button', 'Continue');
    });
    
    it('can select on third page', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/select-your-identity-options-1');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'I do not have any of these types of identity').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options-2');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });
        
        cy.contains('label', 'A bank account').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/your-chosen-identity-options');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('your bank account');        
        cy.contains('button', 'Continue');
    });
});
