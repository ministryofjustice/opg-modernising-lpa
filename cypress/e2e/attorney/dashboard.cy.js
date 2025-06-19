import {randomShareCode, TestEmail} from "../../support/e2e.js";

describe('Attorney dashboard', () => {
    context('original attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Jessie Jones')
        })
    })

    context('replacement attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-replacement')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Blake Buckley')
        })
    })

    context('trust corporation attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-trust-corporation')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('First Choice Trust Corporation Ltd.')
        })
    })

    context('replacement trust corporation attorney', () => {
        it('has a dashboard card', () => {
            cy.visit('/fixtures/attorney?redirect=&progress=signedByCertificateProvide&options=is-trust-corporation&options=is-replacement')

            cy.url().should('contain', '/dashboard')
            cy.checkA11yApp();

            cy.contains('a', 'Go to task list').click()

            cy.url().should('contain', '/task-list')
            cy.checkA11yApp();

            cy.visitLpa('/confirm-your-details')
            cy.contains('Second Choice Trust Corporation Ltd.')
        })
    })

    context('with existing LPAs and an attorney share code', () => {
        it('can add access to act as an attorney on an LPA', () => {
            const randomCode = randomShareCode();
            cy.visit(`/fixtures/attorney?redirect=&withShareCode=${randomCode}&progress=signedByCertificateProvider&email=${TestEmail}`);
            cy.visit(`/fixtures?redirect=/task-list&progress=provideYourDetails`);

            cy.contains('a', 'Make or add an LPA').click();

            cy.contains('a', 'Continue').click();

            cy.contains('label', 'I have a code inviting me to be an attorney').click();
            cy.url().should('contain', '/add-an-lpa');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/attorney-enter-reference-number');
            cy.get('#f-reference-number').invoke('val', randomCode);
            cy.checkA11yApp();

            cy.contains('button', 'Save and continue').click();
            cy.contains('We have identified your attorney access code').click();
        })
    })
})
