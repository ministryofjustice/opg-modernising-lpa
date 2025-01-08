describe('Check your details', () => {
    it('shows my details', () => {
        cy.visit('/fixtures?redirect=/check-your-details&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence');

        cy.checkA11yApp();
        cy.contains('Sam Smith').parent().contains('a', 'Change').click()
        cy.url().should('contain', '/your-name');
        cy.contains('button', 'Save and continue').click();

        cy.contains('2 January 2000').parent().contains('a', 'Change').click()
        cy.url().should('contain', '/your-date-of-birth');
        cy.contains('button', 'Save and continue').click();

        cy.contains('1 RICHMOND PLACE').parent().contains('a', 'Change').click()
        cy.url().should('contain', '/your-address');
        cy.contains('button', 'Save and continue').click();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/we-have-contacted-voucher');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click();

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp();
    });

    it('tells me about a pending payment', () => {
        cy.visit('/fixtures?redirect=/check-your-details&progress=payForTheLpa&feeType=NoFee&paymentTaskProgress=Pending');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/we-have-contacted-voucher');
        cy.checkA11yApp();
        cy.contains('We are processing your LPA fee request');
    });
});
