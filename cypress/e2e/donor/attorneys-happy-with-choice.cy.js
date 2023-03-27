describe('Attorneys happy with choice', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-attorneys-summary&withAttorneys=2&cookiesAccepted=1');

        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        cy.get('input[value=jointly]').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/are-you-happy-if-one-attorney-cant-act-none-can');
        cy.checkA11yApp();
    });

    it('can be answered yes', () => {
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-replacement-attorneys');
    });

    it('can be answered no', () => {
        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/are-you-happy-if-remaining-attorneys-can-continue-to-act');
        cy.checkA11yApp();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/do-you-want-replacement-attorneys');

        cy.go('back');

        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/do-you-want-replacement-attorneys');
    });
});
