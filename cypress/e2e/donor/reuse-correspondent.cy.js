import { randomAccessCode } from "../../support/e2e";

describe('Reuse correspondent', () => {
    before(() => {
        const sub = randomAccessCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=provideYourDetails&redirect=/task-list`);

        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.get('#f-email').invoke('val', 'email@example.com');
        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();

        cy.visit(`/fixtures?donorSub=${sub}&progress=provideYourDetails&redirect=/task-list`);
    });

    it('selects a previously entered correspondent', () => {
        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('label', 'Select John Smith').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('You’ve added a correspondent');
        cy.contains('John Smith');
        cy.contains('email@example.com');
    });
});
