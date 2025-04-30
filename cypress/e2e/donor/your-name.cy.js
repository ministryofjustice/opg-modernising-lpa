describe('Your name', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-first-names').invoke('val', 'John');
            cy.get('#f-last-name').invoke('val', 'Doe');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

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
            cy.get('#f-other-names').invoke('val', 'c'.repeat(51));

            cy.contains('button', 'Save and continue').click();

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
            cy.contains('[for=f-other-names] + div + .govuk-error-message', 'Other names you are known by must be 50 characters or less');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('warns when name shared with other actor', () => {
            cy.get('#f-first-names').invoke('val', 'Jessie');
            cy.get('#f-last-name').invoke('val', 'Jones');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-name');

            cy.contains('You have already entered Jessie Jones as an attorney on your LPA.');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });
    });
});
