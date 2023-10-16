import { TestEmail } from "../../support/e2e";

describe.skip('Choose attorneys task', () => {
    it('is not started when no attorneys are set', () => {
        cy.visit('/fixtures?redirect=/task-list');

        cy.contains('a', 'Choose your attorneys').parent().parent().contains('Not started');
    });

    it('is in progress if I start adding an attorney', () => {
        cy.visit('/fixtures?redirect=/task-list');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('In progress (1)');
    });

    it('is completed if enter an attorneys details', () => {
        cy.visit('/fixtures?redirect=/task-list');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('Completed (1)');
    });

    it('is completed if enter an attorneys details using address', () => {
        cy.visit('/fixtures?redirect=/task-list');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('In progress (1)');
        cy.go('back');

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();
        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('Completed (1)');
    });

    it('is in progress if I enter multiple attorneys details', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('In progress (2)');
    });

    it('is completed if I enter multiple attorneys details with how they act', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        cy.contains('button', 'Skip').click();

        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        cy.get('input[value=jointly-and-severally]').click();
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('Completed (2)');
    });

    it('is completed if I enter multiple attorneys details when jointly', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
        cy.contains('a', 'Choose your attorneys').click();
        cy.contains('a', 'Continue').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        cy.contains('button', 'Skip').click();

        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        cy.get('input[value=jointly]').click();
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list');
        cy.contains('a', 'Choose your attorneys').parent().parent().contains('Completed (2)');
    });
});
