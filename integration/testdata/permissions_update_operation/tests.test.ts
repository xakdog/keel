import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";
import { PostType } from "@teamkeel/sdk";
import { isNullishCoalesce } from "typescript";
import { listenerCount } from "process";

beforeEach(resetDatabase);

test("string permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: { value: "goodbye" } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: { value: "hello" } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.updateWithTextPermissionLiteral({
      where: { id: post.id },
      values: { title: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("number permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: { value: 100 } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: { value: 1 } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.updateWithNumberPermissionLiteral({
      where: { id: post.id },
      values: { views: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("boolean permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: { value: false } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: { value: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on literal - null value - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.updateWithBooleanPermissionLiteral({
      where: { id: post.id },
      values: { active: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("enum permission on literal - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: { value: PostType.Lifestyle } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on literal - not matching value - is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.updateWithEnumPermissionLiteral({
      where: { id: post.id },
      values: { type: { value: PostType.Lifestyle } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

// test("enum permission on literal - null value - is not authorized", async () => {
//   const post = await actions.createWithEnum({ type: null });

//   await expect(
//     actions.updateWithEnumPermissionLiteral({
//       where: { id: post.id },
//       values: { type: null },
//     })
//   ).toHaveAuthorizationError();

//   // Ensure the update did not complete
//   const samePost = await models.post.findOne({ id: post.id });
//   expect(samePost!.type).equals(null);
// });

test("string permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: { value: "goodbye" } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: { value: "hello" } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("string permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithText({ title: { value: "goodbye" } });

  await expect(
    actions.updateWithTextPermissionFromField({
      where: { id: post.id },
      values: { title: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.title).equals("goodbye");
});

test("number permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 1 } });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: { value: 100 } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on field - not matching value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: { value: 1 } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("number permission on field - null value - is not authorized", async () => {
  const post = await actions.createWithNumber({ views: { value: 100 } });

  await expect(
    actions.updateWithNumberPermissionFromField({
      where: { id: post.id },
      values: { views: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.views).equals(100);
});

test("boolean permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: true } });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: { value: false } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on field - field is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: { value: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("boolean permission on field - null - is not authorized", async () => {
  const post = await actions.createWithBoolean({ active: { value: false } });

  await expect(
    actions.updateWithBooleanPermissionFromField({
      where: { id: post.id },
      values: { active: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.active).equals(false);
});

test("enum permission on field - matching value - is authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Technical },
  });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: { value: PostType.Lifestyle } },
    })
  ).not.toHaveAuthorizationError();

  // Ensure the update completed
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on field - field is not authorized", async () => {
  const post = await actions.createWithEnum({
    type: { value: PostType.Lifestyle },
  });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: { value: PostType.Technical } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(PostType.Lifestyle);
});

test("enum permission on field - null - is not authorized", async () => {
  const post = await actions.createWithEnum({ type: { isNull: true } });

  await expect(
    actions.updateWithEnumPermissionFromField({
      where: { id: post.id },
      values: { type: { isNull: true } },
    })
  ).toHaveAuthorizationError();

  // Ensure the update did not complete
  const samePost = await models.post.findOne({ id: post.id });
  expect(samePost!.type).equals(null);
});

test("identity permission - correct identity in context - is authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.withAuthToken(token).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: { value: "hello" } },
    })
  ).not.toHaveAuthorizationError();
});

test("identity permission - incorrect identity in context - is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user1@keel.xyz",
      password: "1234",
    },
  });

  const { token: token2 } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user2@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.withAuthToken(token2).updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: { value: "hello" } },
    })
  ).toHaveAuthorizationError();
});

test("identity permission - no identity in context - is not authorized", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const post = await actions.withAuthToken(token).createWithIdentity({});

  await expect(
    actions.updateWithIdentityPermission({
      where: { id: post.id },
      values: { title: { value: "hello" } },
    })
  ).toHaveAuthorizationError();
});

test("true value permission - unauthenticated identity - is authorized", async () => {
  const post = await actions.createWithText({ title: { value: "hello" } });

  await expect(
    actions.updateWithTrueValuePermission({
      where: { id: post.id },
      values: { title: { value: "hello again" } },
    })
  ).not.toHaveAuthorizationError();
});
