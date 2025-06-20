import { useState, useEffect } from "react";
import { useServices } from "../../contexts/ServiceContext";
import { COLLECTION_TYPES, COLLECTION_TYPE_LABELS } from "../../constants";
import Input from "../ui/Input";
import Select from "../ui/Select";
import Button from "../ui/Button";

/**
 * Collection Form component for creating and editing collections
 * Follows Single Responsibility Principle - only handles collection form logic
 */
const CollectionForm = ({
  collection = null,
  onSubmit,
  onCancel,
  loading = false,
}) => {
  const { cryptoService } = useServices();
  const [formData, setFormData] = useState({
    name: "",
    type: COLLECTION_TYPES.FOLDER,
    parent_id: "",
  });
  const [errors, setErrors] = useState({});

  // Initialize form data when editing
  useEffect(() => {
    if (collection) {
      try {
        const decryptedName = cryptoService.decrypt(collection.encrypted_name);
        setFormData({
          name: decryptedName,
          type: collection.collection_type,
          parent_id: collection.parent_id || "",
        });
      } catch (error) {
        console.error("Failed to decrypt collection name:", error);
        setFormData({
          name: "",
          type: collection.collection_type,
          parent_id: collection.parent_id || "",
        });
      }
    }
  }, [collection, cryptoService]);

  const validateForm = () => {
    const newErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = "Collection name is required";
    } else if (formData.name.trim().length < 2) {
      newErrors.name = "Collection name must be at least 2 characters";
    }

    if (!formData.type) {
      newErrors.type = "Collection type is required";
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
      name: formData.name.trim(),
      type: formData.type,
      parent_id: formData.parent_id || null,
    };

    // Add version for updates
    if (collection) {
      submitData.version = collection.version || 1;
    }

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

  const typeOptions = Object.entries(COLLECTION_TYPE_LABELS).map(
    ([value, label]) => ({
      value,
      label,
    }),
  );

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="Collection Name"
        type="text"
        value={formData.name}
        onChange={(e) => handleChange("name", e.target.value)}
        error={errors.name}
        required
        placeholder="Enter collection name"
        autoFocus
      />

      <Select
        label="Collection Type"
        value={formData.type}
        onChange={(e) => handleChange("type", e.target.value)}
        options={typeOptions}
        error={errors.type}
        required
      />

      {/* TODO: Add parent collection selector in future iteration */}
      {formData.parent_id && (
        <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
          <p className="text-sm text-blue-800">
            This collection will be nested under the selected parent.
          </p>
        </div>
      )}

      <div className="flex space-x-3 pt-4">
        <Button
          type="submit"
          loading={loading}
          disabled={loading}
          className="flex-1"
        >
          {collection ? "Update Collection" : "Create Collection"}
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

export default CollectionForm;
