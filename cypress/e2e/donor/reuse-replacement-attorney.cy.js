import { randomAccessCode } from "../../support/e2e";

describe('Reuse replacement attorney', () => {
    beforeEach(() => {
        const sub = randomAccessCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourAttorneys&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourAttorneys&attorneys=trust-corporation&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourAttorneys&attorneys=single&redirect=/do-you-want-replacement-attorneys`);
    });

    it('can select a previously entered attorney', () => {
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('label', 'Select Robin Redcar').click();
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('You have added 1 replacement attorney');
        cy.contains('Robin Redcar');
    });

    it('can select a previously enter trust corporation', () => {
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();
        cy.contains('a', 'My replacement attorney is a trust corporation').click();

        cy.checkA11yApp();
        cy.contains('label', 'Select First Choice Trust Corporation Ltd.').click();;
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('You have added 1 replacement attorney');
        cy.contains('First Choice Trust Corporation Ltd.');
    });

    it('can enter a new attorney', () => {
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();

        cy.url().should('include', '/enter-replacement-attorney');
    });
});
