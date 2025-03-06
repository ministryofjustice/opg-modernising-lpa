describe('Confirm your name', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?progress=notStarted&redirect=/confirm-your-name');
    });

    it('shows my name', () => {
        cy.checkA11yApp();

        cy.contains('Vivian');
        cy.contains('Vaughn');

        cy.contains('button', 'Continue').click();
        cy.get('ul li:first-child').should('contain', 'Completed');
    });

    it('can update my name', () => {
        cy.contains('div', 'Vivian').contains('a', 'Change').click();

        cy.url().should('contain', '/your-name')
        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', 'Barry');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/confirm-your-name')
        cy.contains('Barry');

        cy.contains('button', 'Continue').click();
        cy.get('ul li:first-child').should('contain', 'Completed');
    });

    it('warns when last name matches donor', () => {
        cy.contains('div', 'Vivian').contains('a', 'Change').click();

        cy.url().should('contain', '/your-name')
        cy.checkA11yApp();
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.contains('button', 'Save and continue').click();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.checkA11yApp();
        cy.contains('surname matches your surname');

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.get('ul li:first-child').should('contain', 'Completed');
    });

    it('warns when name matches another actor', () => {
        cy.contains('div', 'Vivian').contains('a', 'Change').click();

        cy.url().should('contain', '/your-name')
        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', 'Charlie');
        cy.get('#f-last-name').invoke('val', 'Cooper');
        cy.contains('button', 'Save and continue').click();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.checkA11yApp();
        cy.contains('You have entered a name that matches someone');

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.get('ul li:first-child').should('contain', 'Completed');
    });
});
