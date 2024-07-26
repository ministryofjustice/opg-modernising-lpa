import { AddressFormAssertions, TestEmail } from "../../support/e2e";

describe('Choose replacement attorneys task', () => {
  it('is not started when no replacement attorneys are set', () => {
    cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys');

    cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Not started');
  });

  it('is completed if I do not want replacement attorneys', () => {
    cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys');

    cy.contains('a', 'Choose your replacement attorneys').click();

    cy.contains('label', 'No').click();
    cy.contains('button', 'Save and continue').click();

    cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Completed');
  });

  it('is in progress if I do want replacement attorneys', () => {
    cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys');
    cy.contains('a', 'Choose your replacement attorneys').click();

    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.visitLpa('/task-list');
    cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('In progress');
  });

  it('is completed if enter a replacement attorneys details', () => {
    cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
    cy.contains('a', 'Choose your replacement attorneys').click();

    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.get('#f-first-names').type('John');
    cy.get('#f-last-name').type('Doe');
    cy.get('#f-email').type(TestEmail);
    cy.get('#f-date-of-birth').type('1');
    cy.get('#f-date-of-birth-month').type('2');
    cy.get('#f-date-of-birth-year').type('1990');
    cy.contains('button', 'Save and continue').click();

    cy.contains('label', 'Enter a new address').click();
    cy.contains('button', 'Continue').click();
    AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

    cy.visitLpa('/task-list');
    cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
      cy.contains('Completed');
      cy.contains('1 added');
    });
  });

  it('is in progress if enter a replacement attorneys details then add attorneys', () => {
    cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
    cy.contains('a', 'Choose your replacement attorneys').click();

    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.get('#f-first-names').type('John');
    cy.get('#f-last-name').type('Doe');
    cy.get('#f-email').type(TestEmail);
    cy.get('#f-date-of-birth').type('1');
    cy.get('#f-date-of-birth-month').type('2');
    cy.get('#f-date-of-birth-year').type('1990');
    cy.contains('button', 'Save and continue').click();

    cy.visitLpa('/task-list');

    cy.contains('a', 'Choose your attorneys').click();
    cy.contains('button', 'Continue').click();

    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Continue').click();

    cy.get('#f-first-names').type('Janet');
    cy.get('#f-last-name').type('Doe');
    cy.get('#f-email').type(TestEmail);
    cy.get('#f-date-of-birth').type('1');
    cy.get('#f-date-of-birth-month').type('2');
    cy.get('#f-date-of-birth-year').type('1990');
    cy.contains('button', 'Save and continue').click();

    cy.contains('label', 'Enter a new address').click();
    cy.contains('button', 'Continue').click();
    AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

    cy.contains('label', 'No').click();
    cy.contains('button', 'Continue').click();

    cy.get('input[value=jointly-and-severally]').click();
    cy.contains('button', 'Save and continue').click();

    cy.visitLpa('/task-list');
    cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
      cy.contains('In progress');
      cy.contains('1 added');
    });
  });

  describe('having a single attorney and a single replacement attorney', () => {
    it('is completed', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=single');
      cy.contains('a', 'Choose your replacement attorneys').click();

      cy.contains('label', 'Yes').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('#f-first-names').type('John');
      cy.get('#f-last-name').type('Doe');
      cy.get('#f-email').type(TestEmail);
      cy.get('#f-date-of-birth').type('1');
      cy.get('#f-date-of-birth-month').type('2');
      cy.get('#f-date-of-birth-year').type('1990');
      cy.contains('button', 'Save and continue').click();

      cy.contains('label', 'Enter a new address').click();
      cy.contains('button', 'Continue').click();
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });
  });

  describe('having a single attorney and multiple replacement attorneys', () => {
    beforeEach(() => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourReplacementAttorneys&attorneys=single&replacementAttorneys=single');
      cy.contains('a', 'Choose your replacement attorneys').click();

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
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();
    });

    it('is in progress', () => {
      cy.visitLpa('/task-list');

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('In progress');
        cy.contains('2 added');
      });
    });

    it('is completed if replacements act jointly and severally', () => {
      cy.get('input[value=jointly-and-severally]').click();
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if replacement act jointly', () => {
      cy.get('input[value=jointly]').click();
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if replacement act mixed', () => {
      cy.get('input[value=jointly-for-some-severally-for-others]').click();
      cy.get('textarea').type('Some details');
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });
  });

  describe('having jointly and severally attorneys and a single replacement attorney', () => {
    beforeEach(() => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys');
      cy.contains('a', 'Choose your replacement attorneys').click();

      cy.contains('label', 'Yes').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('#f-first-names').type('John');
      cy.get('#f-last-name').type('Doe');
      cy.get('#f-email').type(TestEmail);
      cy.get('#f-date-of-birth').type('1');
      cy.get('#f-date-of-birth-month').type('2');
      cy.get('#f-date-of-birth-year').type('1990');
      cy.contains('button', 'Save and continue').click();

      cy.contains('label', 'Enter a new address').click();
      cy.contains('button', 'Continue').click();
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();
    });

    it('is completed if step in as soon as one', () => {
      cy.contains('label', 'All together, as soon as one').click();
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });

    it('is completed if step in when none', () => {
      cy.contains('label', 'All together, when none').click();
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });

    it('is completed if step in some other way', () => {
      cy.contains('label', 'In a particular order').click();
      cy.get('textarea').type('Details');
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });
  });

  describe('having jointly attorneys and a single replacement attorney', () => {
    it('is completed', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=jointly');
      cy.contains('a', 'Choose your replacement attorneys').click();

      cy.contains('label', 'Yes').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('#f-first-names').type('John');
      cy.get('#f-last-name').type('Doe');
      cy.get('#f-email').type(TestEmail);
      cy.get('#f-date-of-birth').type('1');
      cy.get('#f-date-of-birth-month').type('2');
      cy.get('#f-date-of-birth-year').type('1990');
      cy.contains('button', 'Save and continue').click();

      cy.contains('label', 'Enter a new address').click();
      cy.contains('button', 'Continue').click();
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });
  });

  describe('having jointly for some attorneys and a single replacement attorney', () => {
    it('is completed', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourAttorneys&attorneys=jointly-for-some-severally-for-others');
      cy.contains('a', 'Choose your replacement attorneys').click();

      cy.contains('label', 'Yes').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('#f-first-names').type('John');
      cy.get('#f-last-name').type('Doe');
      cy.get('#f-email').type(TestEmail);
      cy.get('#f-date-of-birth').type('1');
      cy.get('#f-date-of-birth-month').type('2');
      cy.get('#f-date-of-birth-year').type('1990');
      cy.contains('button', 'Save and continue').click();

      cy.contains('label', 'Enter a new address').click();
      cy.contains('button', 'Continue').click();
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('1 added');
      });
    });
  });

  describe('having jointly and severally attorneys and multiple replacement attorneys', () => {
    beforeEach(() => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourReplacementAttorneys&replacementAttorneys=single');
      cy.contains('a', 'Choose your replacement attorneys').click();

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
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();
    });

    it('is completed if step in as soon as one', () => {
      cy.contains('label', 'All together, as soon as one').click();
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is in progress if step in when none', () => {
      cy.contains('label', 'All together, when none').click();
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('In progress');
        cy.contains('2 added');
      });
    });

    it('is completed if step in when none and jointly and severally', () => {
      cy.contains('label', 'All together, when none').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('input[value=jointly-and-severally]').click();
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if step in when none and jointly', () => {
      cy.contains('label', 'All together, when none').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('input[value=jointly]').click();
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if step in when none and mixed', () => {
      cy.contains('label', 'All together, when none').click();
      cy.contains('button', 'Save and continue').click();

      cy.get('input[value=jointly-for-some-severally-for-others]').click();
      cy.get('textarea').type('Some details');
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if in some other way', () => {
      cy.contains('label', 'In a particular order').click();
      cy.get('textarea').type('Details');
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });
  });

  describe('having jointly attorneys and multiple replacement attorneys', () => {
    beforeEach(() => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourReplacementAttorneys&attorneys=jointly&replacementAttorneys=single');
      cy.contains('a', 'Choose your replacement attorneys').click();

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
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();
    });

    it('is completed if jointly and severally', () => {
      cy.get('input[value=jointly-and-severally]').click();
      cy.contains('button', 'Save and continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if jointly', () => {
      cy.get('input[value=jointly]').click();
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });

    it('is completed if mixed', () => {
      cy.get('input[value=jointly-for-some-severally-for-others]').click();
      cy.get('textarea').type('Some details');
      cy.contains('button', 'Save and continue').click();

      cy.visitLpa('/task-list');
      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });
  });

  describe('having jointly for some attorneys and multiple replacement attorneys', () => {
    it('is completed', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=chooseYourReplacementAttorneys&attorneys=jointly-for-some-severally-for-others&replacementAttorneys=single');
      cy.contains('a', 'Choose your replacement attorneys').click();

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
      AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();

      cy.contains('a', 'Choose your replacement attorneys').parent().parent().within(() => {
        cy.contains('Completed');
        cy.contains('2 added');
      });
    });
  });
});
