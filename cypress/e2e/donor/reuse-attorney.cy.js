import { randomShareCode } from "../../support/e2e";

describe('Reuse attorney', () => {
    beforeEach(() => {
        const sub = randomShareCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourAttorneys&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourAttorneys&attorneys=trust-corporation&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=provideYourDetails&redirect=/choose-attorneys-guidance`);
    });

    it('can select a previously entered attorney', () => {
        cy.contains('a', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('label', 'Select Robin Redcar').click();
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('You have added 1 attorney');
        cy.contains('Robin Redcar');
    });

    it('can select a previously enter trust corporation', () => {
        cy.contains('a', 'Continue').click();
        cy.contains('a', 'My attorney is a trust corporation').click();

        cy.checkA11yApp();
        cy.contains('label', 'Select First Choice Trust Corporation Ltd.').click();;
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('You have added 1 attorney');
        cy.contains('First Choice Trust Corporation Ltd.');
    });

    it('can enter a new attorney', () => {
        cy.contains('a', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();

        cy.url().should('include', '/enter-attorney');
    });
});
