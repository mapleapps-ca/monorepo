import { useState } from "react";
import { PERMISSION_LEVELS, PERMISSION_LABELS } from "../../constants";
import { isValidEmail } from "../../utils";
import Input from "../ui/Input";
import Select from "../ui/Select";
import Button from "../ui/Button";

/**
 * Share Collection Form component
 * Follows Single Responsibility Principle - only handles collection sharing form logic
 */
const ShareCollectionForm = ({ onSubmit, onCancel, loading = false }) => {
  const [formData, setFormData] = useState({
    recipient_email: "",
    permission_level: PERMISSION_LEVELS.READ_ONLY,
    share_with_descendants: true,
  });
  const [errors, setErrors] = useState({});

  const validateForm = () => {
    const newErrors = {};

    if (!formData.recipient_email.trim()) {
      newErrors.recipient_email = "Email is required";
    } else if (!isValidEmail(formData.recipient_email)) {
      newErrors.recipient_email = "Please enter a valid email address";
    }

    if (!formData.permission_level) {
      newErrors.permission_level = "Permission level is required";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    const submitData = {
      recipient_email: formData.recipient_email.trim(),
      permission_level: formData.permission_level,
      share_with_descendants: formData.share_with_descendants,
      // In a real app, you'd need to get the recipient's public key
      // and the collection key to encrypt for sharing
      recipient_id: "generated-user-id", // This would come from user lookup
    };

    onSubmit(submitData);
  };

  const handleChange = (field, value) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }
  };

  const permissionOptions = Object.entries(PERMISSION_LABELS).map(
    ([value, label]) => ({
      value,
      label,
    }),
  );

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Email Address"
        type="email"
        value={formData.recipient_email}
        onChange={(e) => handleChange("recipient_email", e.target.value)}
        error={errors.recipient_email}
        required
        placeholder="Enter user's email address"
        autoFocus
      />

      <Select
        label="Permission Level"
        value={formData.permission_level}
        onChange={(e) => handleChange("permission_level", e.target.value)}
        options={permissionOptions}
        error={errors.permission_level}
        required
      />

      <div className="flex items-center space-x-2">
        <input
          type="checkbox"
          id="share_descendants"
          checked={formData.share_with_descendants}
          onChange={(e) =>
            handleChange("share_with_descendants", e.target.checked)
          }
          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        />
        <label htmlFor="share_descendants" className="text-sm text-gray-700">
          Share all nested collections
        </label>
      </div>

      <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
        <p className="text-sm text-yellow-800">
          <strong>Permission Levels:</strong>
        </p>
        <ul className="text-xs text-yellow-700 mt-1 space-y-1">
          <li>
            <strong>Read Only:</strong> Can view collection and files
          </li>
          <li>
            <strong>Read & Write:</strong> Can add/modify files and
            subcollections
          </li>
          <li>
            <strong>Admin:</strong> Full control including sharing and deletion
          </li>
        </ul>
      </div>

      <div className="flex space-x-3 pt-4">
        <Button
          type="submit"
          loading={loading}
          disabled={loading}
          className="flex-1"
        >
          Share Collection
        </Button>

        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={loading}
        >
          Cancel
        </Button>
      </div>
    </form>
  );
};

export default ShareCollectionForm;
