const { TestMobile, TestEmail, randomShareCode } = require("../../support/e2e");

describe('As a trust corporation', () => {
    beforeEach(() => {
        const shareCode = randomShareCode()
        cy.visit(`/fixtures/attorney?redirect=/attorney-start&is-trust-corporation=1&progress=readTheLPA&&withShareCode=${shareCode}&email=${TestEmail}`);

        // start
        cy.contains('a', 'Start').click();
        cy.contains('label', 'Random value').click();
        cy.get('form').submit();

        // enter reference number
        cy.get('#f-reference-number').type(shareCode);
        cy.contains('button', 'Save and continue').click();

        // acting as an attorney
        cy.contains('We have identified the trust corporation’s attorney reference number');
        cy.contains('a', 'Continue').click();

        // task list
        cy.contains('a', 'Confirm your details').click();

        // phone number
        cy.get('#f-phone').type(TestMobile);
        cy.contains('button', 'Save and continue').click();

        // language preferences
        cy.get('[name="language-preference"]').check('cy', { force: true })
        cy.contains('button', 'Save and continue').click()

        // confirm your company details
        cy.contains('07700 900 000');
        cy.contains('Welsh');
        cy.contains('Confirm your company details');
        cy.contains('First Choice Trust Corporation Ltd.');
        cy.contains('555555555');
        cy.contains('simulate-delivered@notifications.service.gov.uk');
        cy.contains('2 RICHMOND PLACE');
        cy.contains('B14 7ED');
        cy.contains('button', 'Continue').click();

        // task list
        cy.contains('Read the LPA').click();
        cy.contains('button', 'Continue').click();

        // legal rights and responsibilities
        cy.contains('Sign the LPA').click();
        cy.contains('Before signing, you must read the trust corporation’s legal rights and responsibilities as an attorney.');
        cy.contains('a', 'Continue').click();

        // what happens when you sign the lpa
        cy.contains('What happens when you sign the LPA');
        cy.contains('a', 'Continue to signing page').click();
    });

    it('allows a single signatory', () => {
        // sign
        cy.contains('Sign the LPA on behalf of the trust corporation');
        cy.get('#f-first-names').type('Sign');
        cy.get('#f-last-name').type('Signson');
        cy.get('#f-professional-title').type('Pro signer');
        cy.get('#f-confirm').check({ force: true });
        cy.contains('button', 'Submit signature').click();

        // would like a 2nd signatory
        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        // what happens next
        cy.contains('First Choice Trust Corporation Ltd. has formally agreed to be an attorney');
        cy.contains('a', 'Go to your dashboard');
    });

    it('allows a second signatory', () => {
        // sign
        cy.contains('Sign the LPA on behalf of the trust corporation');
        cy.get('#f-first-names').type('Sign');
        cy.get('#f-last-name').type('Signson');
        cy.get('#f-professional-title').type('Pro signer');
        cy.get('#f-confirm').check({ force: true });
        cy.contains('button', 'Submit signature').click();

        // would like a 2nd signatory
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        // task list
        cy.visitLpa('/task-list');
        cy.contains('Sign the LPA (signatory 1)');
        cy.contains('Sign the LPA (signatory 2)').click();

        // sign
        cy.get('#f-first-names').type('Sign2');
        cy.get('#f-last-name').type('Signson2');
        cy.get('#f-professional-title').type('Pro signer2');
        cy.get('#f-confirm').check({ force: true });
        cy.contains('button', 'Submit signature').click();

        // what happens next
        cy.contains('First Choice Trust Corporation Ltd. has formally agreed to be an attorney');
        cy.contains('a', 'Go to your dashboard');
    });

    it('can remove second signatory', () => {
        // sign
        cy.contains('Sign the LPA on behalf of the trust corporation');
        cy.get('#f-first-names').type('Sign');
        cy.get('#f-last-name').type('Signson');
        cy.get('#f-professional-title').type('Pro signer');
        cy.get('#f-confirm').check({ force: true });
        cy.contains('button', 'Submit signature').click();

        // would like a 2nd signatory
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        // task list
        cy.visitLpa('/task-list');
        cy.contains('Sign the LPA (signatory 1)');
        cy.contains('Sign the LPA (signatory 2)').click();

        // sign
        cy.contains('a', 'The trust corporation no longer requires a second signatory').click();

        // would like a 2nd signatory
        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        // what happens next
        cy.contains('First Choice Trust Corporation Ltd. has formally agreed to be an attorney');
        cy.contains('a', 'Go to your dashboard');
    });
});
