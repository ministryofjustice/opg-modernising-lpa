describe('Your independent witness', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/your-independent-witness');
    });

    it('can be submitted', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-independent-witness-mobile');
    });

    it('errors when empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
    });

    it('errors when names too long', () => {
        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));

        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
    });

    it('warns when name shared with other actor', () => {
        cy.visit('/fixtures?redirect=/your-independent-witness&progress=chooseYourAttorneys');

        cy.get('#f-first-names').type('Jessie');
        cy.get('#f-last-name').type('Jones');
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-independent-witness');

        cy.contains('There is also an attorney called Jessie Jones.');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-independent-witness-mobile');
    });
});
